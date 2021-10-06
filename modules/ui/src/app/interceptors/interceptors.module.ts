import { CommonModule } from '@angular/common';
import { HTTP_INTERCEPTORS } from '@angular/common/http';
import { NgModule } from '@angular/core';
import { ErrorInterceptor } from '@app/interceptors/error.interceptor';
import { TokenInterceptor } from '@app/interceptors/token.interceptor';

@NgModule({
    declarations: [],
    imports: [
        CommonModule
    ],
    providers: [
        {
            provide: HTTP_INTERCEPTORS,
            useClass: TokenInterceptor,
            multi: true,
        },
        {
            provide: HTTP_INTERCEPTORS,
            useClass: ErrorInterceptor,
            multi: true,
        },
    ],
})
export class InterceptorsModule { }
