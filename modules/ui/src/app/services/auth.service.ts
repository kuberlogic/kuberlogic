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
