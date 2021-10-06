import { Component, OnInit } from '@angular/core';
import { AuthService } from '@services/auth.service';
import { environment } from '@environments/environment';

@Component({
    selector: 'kl-header',
    templateUrl: './header.component.html',
    styleUrls: ['./header.component.scss'],
})
export class HeaderComponent implements OnInit {

    constructor(
        private authService: AuthService,
    ) { }

    ngOnInit(): void {
    }

    onLogout(): void {
        this.authService.logout();
    }

    getConsoleMonitoringUrl() {
        return `${environment.monitoringConsoleUrl}?token=${this.authService.getToken()}`;
    }
}
