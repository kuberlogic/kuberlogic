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

import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { environment } from '@environments/environment';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';

const LOCAL_STORAGE_TOKEN = 'KUBERLOGIC_TOKEN';

@Injectable({
    providedIn: 'root'
})
export class AuthService {
    isLoggedIn = false;

    constructor(
        private router: Router,
        private http: HttpClient,
    ) {}

    login(username: string, password: string): Observable<void> {
        return this.http.post<any>(`${environment.apiUrl}/login`, {username, password})
            .pipe(map((result: any) => {
                this.saveToken(result.access_token);
                this.router.navigate(['/', 'services']);
            }));
    }

    canActivate(): boolean {
        if (this.isLoggedIn) {
            this.router.navigate(['/']);
            return false;
        }
        return true;
    }

    saveToken(token: string): void {
        localStorage.setItem(LOCAL_STORAGE_TOKEN, token);
        this.isLoggedIn = true;
    }

    logout(): void {
        this.isLoggedIn = false;
        localStorage.removeItem(LOCAL_STORAGE_TOKEN);
        this.router.navigate(['/', 'login']);
    }

    getToken(): string | null {
        const token = localStorage.getItem(LOCAL_STORAGE_TOKEN);
        this.isLoggedIn = token !== null;
        return token;
    }
}
