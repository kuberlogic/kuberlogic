import { ChangeDetectionStrategy, Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { ServiceModel } from '@models/service.model';
import { ServicesPageService } from '@services/services-page.service';
import { Observable } from 'rxjs';

export interface NavLink {
    label: string;
    link: string;
}

@Component({
    selector: 'kl-edit-service',
    templateUrl: './view-service.component.html',
    styleUrls: ['./view-service.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ViewServiceComponent implements OnInit {
    navLinks: NavLink[];
    currentService$!: Observable<ServiceModel | undefined>;

    constructor(
        private route: ActivatedRoute,
        private servicesPageService: ServicesPageService,
    ) {
        this.navLinks = [
            {
                label: 'Connection',
                link: 'connection',
            },
            {
                label: 'Settings',
                link: 'settings',
            },
            {
                label: 'Logs',
                link: 'logs',
            },
            {
                label: 'Backups',
                link: 'backups',
            },
        ];
    }

    ngOnInit(): void {
        this.currentService$ = this.servicesPageService.getService(this.route.snapshot.params.id);
    }

}
