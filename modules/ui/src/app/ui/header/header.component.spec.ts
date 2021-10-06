import { ComponentFixture, TestBed } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';

import { HttpClientTestingModule } from '@angular/common/http/testing';
import { AuthService } from '@services/auth.service';
import { MessagesService } from '@services/messages.service';
import { MockAuthService } from '@testing/mock-auth-service';
import { MockMessageService } from '@testing/mock-messages-service';
import { HeaderComponent } from './header.component';

describe('HeaderComponent', () => {
    let component: HeaderComponent;
    let fixture: ComponentFixture<HeaderComponent>;
    let authService: MockAuthService;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                RouterTestingModule,
                HttpClientTestingModule,
            ],
            declarations: [HeaderComponent],
            providers: [
                { provide: MessagesService, useClass: MockMessageService },
                { provide: AuthService, useClass: MockAuthService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(HeaderComponent);
        component = fixture.componentInstance;
        // @ts-ignore
        authService = TestBed.inject(AuthService);
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should call authService on logout', () => {
        const spy = spyOn(authService, 'logout');
        component.onLogout();
        fixture.detectChanges();

        expect(spy).toHaveBeenCalled();
    });
});
