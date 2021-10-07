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

import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { FormContainerMixin } from '@app/mixins/form-container.mixin';
import { BaseObject } from '@app/mixins/mixins';
import { AuthService } from '@services/auth.service';
import { EMPTY } from 'rxjs';
import { catchError, take } from 'rxjs/operators';

@Component({
    selector: 'kl-login',
    styleUrls: ['./login.component.scss'],
    templateUrl: './login.component.html',
})
export class LoginComponent  extends FormContainerMixin(BaseObject) {
    formGroup: FormGroup;
    serverError = false;
    wrongCredentials = false;

    constructor(
        private fb: FormBuilder,
        private authService: AuthService,
    ) {
        super();
        this.formGroup = this.fb.group({
            password: ['', [Validators.required]],
            username: ['', [Validators.required]],
        });
    }

    onLogin(): void {
        if (this.checkForm()) {
            this.authService.login(
                this.formGroup.controls.username.value,
                this.formGroup.controls.password.value,
            ).pipe(
                catchError((err) => {
                    if (err === 'Unauthorized') {
                        this.wrongCredentials = true;
                    } else {
                        this.serverError = true;
                    }
                    this.formGroup.valueChanges.pipe(take(1)).subscribe(() => {
                        this.serverError = false;
                        this.wrongCredentials = false;
                    });
                    return EMPTY;
                }),
            ).toPromise();
        }
    }
}
