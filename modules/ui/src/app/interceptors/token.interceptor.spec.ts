import { TestBed } from '@angular/core/testing';

import { AuthService } from '@services/auth.service';
import { MockAuthService } from '@testing/mock-auth-service';
import { TokenInterceptor } from './token.interceptor';

describe('TokenInterceptor', () => {
    beforeEach(() => TestBed.configureTestingModule({
        providers: [
            TokenInterceptor,
            { provide: AuthService, useClass: MockAuthService },
        ]
    }));

    it('should be created', () => {
        const interceptor: TokenInterceptor = TestBed.inject(TokenInterceptor);
        expect(interceptor).toBeTruthy();
    });
});
