import { Observable, of } from 'rxjs';

export class MockAuthService {
    login(username: string, password: string): Observable<void> {
        return of();
    }

    canActivate(): boolean {
        return true;
    }

    saveToken(token: string): void {}

    logout(): void {}
    getToken(): string | null {
        return 'token';
    }
}
