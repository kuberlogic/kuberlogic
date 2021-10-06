import { Injectable } from '@angular/core';
import { MatIconRegistry } from '@angular/material/icon';
import { DomSanitizer } from '@angular/platform-browser';

@Injectable({
    providedIn: 'root'
})
export class IconsService {

    constructor(
        private matIconRegistry: MatIconRegistry,
        private domSanitizer: DomSanitizer,
    ) {
    }

    init(): void {
        this.matIconRegistry.addSvgIcon(
            'mysqlIcon',
            this.domSanitizer.bypassSecurityTrustResourceUrl('assets/svg/service-logos/mysql.svg'),
        );
        this.matIconRegistry.addSvgIcon(
            'postgresqlIcon',
            this.domSanitizer.bypassSecurityTrustResourceUrl('assets/svg/service-logos/postgresql.svg'),
        );
    }
}
