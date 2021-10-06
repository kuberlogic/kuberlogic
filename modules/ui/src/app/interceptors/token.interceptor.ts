import {
  HttpEvent,
  HttpHandler,
  HttpInterceptor,
  HttpRequest
} from '@angular/common/http';
import { Injectable } from '@angular/core';
import { AuthService } from '@services/auth.service';
import { Observable } from 'rxjs';

@Injectable()
export class TokenInterceptor implements HttpInterceptor {
    constructor(
        private authService: AuthService,
    ) { }

    intercept(request: HttpRequest<unknown>, next: HttpHandler): Observable<HttpEvent<unknown>> {
        let modifiedRequest = request;
        if (request.url.includes('api') && !request.url.includes('login')) {
            modifiedRequest = request.clone({
                headers: request.headers.set('Authorization', `Bearer ${this.authService.getToken()}`),
            });
        }

        return next.handle(modifiedRequest);
    }
}
