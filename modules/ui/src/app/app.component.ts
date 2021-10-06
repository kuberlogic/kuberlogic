import { Component } from '@angular/core';
import { AuthService } from '@services/auth.service';
import { IconsService } from '@services/icons.service';

@Component({
    selector: 'kl-root',
    templateUrl: './app.component.html',
    styleUrls: ['./app.component.scss']
})
export class AppComponent {
    constructor(
        private icons: IconsService,
        public authService: AuthService,
    ) {
        this.icons.init();
    }

}
