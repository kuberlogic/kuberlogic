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
