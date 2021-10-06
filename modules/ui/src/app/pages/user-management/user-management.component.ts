import { ChangeDetectionStrategy, Component, OnInit } from '@angular/core';

@Component({
    selector: 'kl-user-management',
    templateUrl: './user-management.component.html',
    styleUrls: ['./user-management.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class UserManagementComponent implements OnInit {

    constructor() { }

    ngOnInit(): void {
    }

}
