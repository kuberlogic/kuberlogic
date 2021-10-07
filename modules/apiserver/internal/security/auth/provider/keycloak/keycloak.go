/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package keycloak

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-oidc"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/cache"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/models"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/logging"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/security/auth/policy"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type keycloakAuthProvider struct {
	realmUrl     string
	oidcVerifier *oidc.IDTokenVerifier
	oauthConfig  oauth2.Config

	ctx                context.Context
	cache              cache.Cache
	permissionEnforcer policy.Enforcer

	patToken       *patToken
	securityGrants []string

	httpClient *http.Client
	log        logging.Logger
}

type patToken struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not_before_policy"`
	Scope            string `json:"scope"`

	ReceivedTime time.Time `json:"-"`
}

type protectionApiResourceSet struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	ResourceScopes []string `json:"resource_scopes"`
}

type userPermissions []struct {
	Scopes []string `json:"scopes"`
	Rsid   string   `json:"rsid"`
	Rsname string   `json:"rsname"`
}

const (
	umaGrantType              = "urn:ietf:params:oauth:grant-type:uma-ticket"
	protectionResourceSetType = "kuberlogicservice"
)

func (k *keycloakAuthProvider) GetAuthenticationSecret(username, password string) (string, error) {
	k.log.Debugw("getting authentication secret", "user", username)

	oauth2token, err := k.oauthConfig.PasswordCredentialsToken(k.ctx, username, password)
	if err != nil {
		k.log.Errorw("error getting token", "username", username, "error", err)
		return "", fmt.Errorf("Failed to get token" + err.Error())
	}
	rawIDToken, ok := oauth2token.Extra("id_token").(string)
	if !ok {
		k.log.Debugw("no id_token found in oauth2 token", "oauth2 token", oauth2token)
		return "", fmt.Errorf("no id_token found in oauth2 token")
	}
	idToken, err := k.oidcVerifier.Verify(k.ctx, rawIDToken)
	if err != nil {
		k.log.Errorw("failed to verify ID token", "error", err.Error())
		return "", fmt.Errorf("failed to verify ID token")
	}

	resp := struct {
		Oauth2token   *oauth2.Token
		IDTokenClaims *json.RawMessage
	}{oauth2token, new(json.RawMessage)}

	if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
		k.log.Errorw("error extracting id_token claims from id_token",
			"id_token", idToken, "error", err)
		return "", fmt.Errorf("error extracting id_token claims")
	}
	return resp.Oauth2token.AccessToken, nil
}

func (k *keycloakAuthProvider) Authenticate(token string) (string, string, error) {
	k.log.Debugw("authenticating new user with token")

	p := strings.Split(token, " ")
	if len(p) != 2 {
		k.log.Errorw("error extracting authentication token", "token", token)
		return "", "", fmt.Errorf("error extracting authentication token")
	}
	authToken := p[1]

	idToken, err := k.oidcVerifier.Verify(k.ctx, authToken)
	if err != nil {
		k.log.Errorw("error verifying authentication token", "error", err)
		return "", "", fmt.Errorf("error veryfying authentication token")
	}

	var userInfo struct {
		Username string `json:"preferred_username"`
		Email    string `json:"email"`
	}
	if err := idToken.Claims(&userInfo); err != nil {
		k.log.Errorw("error getting username from authentication token", "error", err)
		return "", "", fmt.Errorf("error getting username from authentication token")
	}

	if userInfo.Username == "" || userInfo.Email == "" {
		return "", "", fmt.Errorf("empty username or email")
	}

	return userInfo.Email, authToken, nil
}

func (k *keycloakAuthProvider) Authorize(principal *models.Principal, action, object string) (bool, error) {
	userKey := principal.Email

	// check cache first
	if permissions, found := k.cache.Get(userKey); found {
		k.log.Debugw("permissions for action on object found in cache",
			"action", action, "object", object)
		authorized, err := k.permissionEnforcer.IsAuthorized(permissions.(policy.Permissions), userKey, object, action)
		return authorized, err
	}

	// get permissions from keycloak
	kPermissions, err := k.getUserPermissions(principal)
	if err != nil {
		k.log.Errorw("error getting permissions from keycloak", "error", err)
		return false, err
	}

	permissions := policy.Permissions{}
	for _, p := range *kPermissions {
		for _, s := range p.Scopes {
			permissions.Rules = append(permissions.Rules, policy.PermissionRule{
				Subject:  userKey,
				Resource: p.Rsname,
				Action:   s,
			})
		}
	}
	k.cache.Set(userKey, permissions, 60)

	authorized, err := k.permissionEnforcer.IsAuthorized(permissions, userKey, object, action)
	return authorized, err
}

func (k *keycloakAuthProvider) CreatePermissionResource(obj string) error {
	if err := k.patTokenRefresh(false); err != nil {
		k.log.Errorw("error refreshing pat token", "error", err)
		return err
	}
	if err := k.createResourceSet(obj); err != nil {
		k.log.Errorw("error creating permissions resource", "error", err)
		return err
	}
	return nil
}

func (k *keycloakAuthProvider) DeletePermissionResource(obj string) error {
	if err := k.patTokenRefresh(false); err != nil {
		k.log.Errorw("error refreshing pat token", "error", err)
		return err
	}
	id, found, err := k.getResourceSetID(obj)
	if err != nil {
		k.log.Errorw("error getting permission resource id", "error", err)
		return err
	}
	if !found {
		return nil
	}

	err = k.deleteResourceSetByID(id)
	if err != nil {
		k.log.Errorw("error deleting permissions resource", "error", err)
		return err
	}
	return nil
}

func (k *keycloakAuthProvider) getUserPermissions(principal *models.Principal) (*userPermissions, error) {
	data := url.Values{}

	data.Set("grant_type", umaGrantType)
	data.Set("audience", k.oauthConfig.ClientID)
	data.Set("response_mode", "permissions")

	req, err := http.NewRequest("POST", k.oauthConfig.Endpoint.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		k.log.Errorw("error building a client for Keycloak authorization services", "error", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+principal.Token)

	res, err := k.httpClient.Do(req)
	if err != nil {
		k.log.Errorw("error requesting Keycloak permissions", "error", err)
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	k.log.Debugw("Keycloak authorization services response", "response", string(bodyBytes))

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting user permissions, status code is %v message is %s", res.StatusCode, string(bodyBytes))
	}

	permissions := &userPermissions{}
	if err := json.Unmarshal(bodyBytes, permissions); err != nil {
		return nil, fmt.Errorf("error unmarshalling keycloak response: " + err.Error())
	}
	return permissions, nil
}

func (k *keycloakAuthProvider) patTokenRefresh(force bool) error {
	if !force {
		if time.Now().Sub(k.patToken.ReceivedTime).Seconds() < float64(k.patToken.ExpiresIn) {
			return nil
		}
	}
	endpoint := k.realmUrl + "/protocol/openid-connect/token"
	data := url.Values{}
	data.Set("client_id", k.oauthConfig.ClientID)
	data.Set("client_secret", k.oauthConfig.ClientSecret)
	data.Set("grant_type", "client_credentials")

	r, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		k.log.Errorw("error creating pat token request", "error", err)
		return err
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	k.log.Debugw("created pat token request", "url", endpoint)

	res, err := k.httpClient.Do(r)
	if err != nil {
		k.log.Errorw("error making pat token request", "error", err)
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		k.log.Errorw("error reading pat response body", "error", err)
		return err
	}

	k.log.Debugw("done pat token request", "status", res.StatusCode, "response", string(body))
	if res.StatusCode != 200 {
		err := fmt.Errorf("wrong status code: %d", res.StatusCode)
		k.log.Errorw("pat token request error", "error", err, "response", string(body))
		return err
	}

	pat := new(patToken)
	if err := json.Unmarshal(body, pat); err != nil {
		k.log.Errorw("error extracting pat token from keycloak response", "error", err)
		return err
	}
	pat.ReceivedTime = time.Now()
	k.patToken = pat
	k.log.Infow("pat token successfully refreshed")
	return nil
}

func (k *keycloakAuthProvider) getResourceSetID(name string) (string, bool, error) {
	// get id first
	type rsIds []string
	id := rsIds{}

	r, err := http.NewRequest("GET", fmt.Sprintf("%s/authz/protection/resource_set", k.realmUrl), nil)
	if err != nil {
		return "", false, err
	}
	r.Header.Set("Authorization", "Bearer "+k.patToken.AccessToken)

	q := r.URL.Query()
	q.Add("name", name)
	r.URL.RawQuery = q.Encode()
	k.log.Debugw("getting resource_set id for service", "url", r.URL.String())

	res, err := k.httpClient.Do(r)
	if err != nil {
		return "", false, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", false, err
	}

	k.log.Debugw("get resource set by id", "response_code", res.StatusCode, "response_body", string(b))

	if res.StatusCode == 404 {
		return "", false, nil
	}
	if res.StatusCode != 200 {
		return "", false, fmt.Errorf("incorrect response code %d response %s", res.StatusCode, string(b))
	}

	if err := json.Unmarshal(b, &id); err != nil {
		return "", false, err
	}

	if len(id) == 0 {
		return "", false, nil
	}
	if len(id) != 1 {
		return "", false, fmt.Errorf("unknown response %v", id)
	}
	return id[0], false, nil
}

func (k *keycloakAuthProvider) deleteResourceSetByID(id string) error {
	r, err := http.NewRequest("DELETE", k.realmUrl+"/authz/protection/resource_set/"+id, nil)
	if err != nil {
		return err
	}
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", k.patToken.AccessToken))

	res, err := k.httpClient.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != 204 {
		return fmt.Errorf("incorrect status code %d response %s", res.StatusCode, string(b))
	}
	return nil
}

func (k *keycloakAuthProvider) createResourceSet(name string) error {
	p := new(protectionApiResourceSet)
	p.Name = name
	p.Type = protectionResourceSetType
	p.ResourceScopes = k.securityGrants
	jsonData, err := json.Marshal(p)
	if err != nil {
		return err
	}

	r, err := http.NewRequest("POST", k.realmUrl+"/authz/protection/resource_set", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", k.patToken.AccessToken))
	r.Header.Set("Content-Type", "application/json")

	res, err := k.httpClient.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != 201 {
		err := fmt.Errorf("incorrect status code %d response %s", res.StatusCode, string(b))
		return err
	}

	k.log.Debugw("permission resource successfully created")
	return nil
}

func NewKeycloakAuthProvider(clientId, clientSecret, realmName, keycloakUrl string, cache cache.Cache, log logging.Logger, securityGrants []string) (*keycloakAuthProvider, error) {
	configUrl := fmt.Sprintf("%s/auth/realms/%s", keycloakUrl, realmName)

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpClient := &http.Client{Transport: tr}

	ctx := oidc.ClientContext(context.TODO(), httpClient)

	log.Debugw("initializing oidc provider with url", "url", configUrl)
	provider, err := oidc.NewProvider(ctx, configUrl)
	if err != nil {
		return nil, fmt.Errorf("error initializing keycloak oidc config: " + err.Error())
	}

	log.Debugw("initializing oauth2 config with client_id", "client_id", clientId)
	oauth2Config := oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  "",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	oidcConfig := &oidc.Config{
		ClientID: clientId,
	}

	// we use external rule enforcer because of limited Ketcloak enforcer
	log.Debugw("initializing permission policy enforcer")
	enforcer := policy.NewEnforcer(cache, log)

	kc := &keycloakAuthProvider{
		realmUrl:     configUrl,
		oidcVerifier: provider.Verifier(oidcConfig),
		oauthConfig:  oauth2Config,

		permissionEnforcer: enforcer,

		securityGrants: securityGrants,

		httpClient: httpClient,

		ctx:   ctx,
		cache: cache,
		log:   log,
	}

	if err := kc.patTokenRefresh(true); err != nil {
		return nil, err
	}
	return kc, nil
}
