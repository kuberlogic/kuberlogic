import { ChangeDetectionStrategy, Component, Inject, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { FormContainerMixin } from '@app/mixins/form-container.mixin';
import { BaseObject } from '@app/mixins/mixins';
import { ServiceDatabaseModel } from '@models/service-database.model';
import { MessagesService } from '@services/messages.service';
import { ServiceDatabasesService } from '@services/service-databases.service';
import { ServiceRestoresService } from '@services/service-restores.service';
import { ServicesPageService } from '@services/services-page.service';
import { Observable } from 'rxjs';

@Component({
    selector: 'kl-restore-service-backups-dialog',
    templateUrl: './restore-service-backups-dialog.component.html',
    styleUrls: ['./restore-service-backups-dialog.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class RestoreServiceBackupsDialogComponent
    extends FormContainerMixin(BaseObject)
    implements OnInit {
    formGroup: FormGroup;
    serviceDatabases$!: Observable<ServiceDatabaseModel[] | undefined>;

    constructor(
        private fb: FormBuilder,
        private messages: MessagesService,
        private servicesPageService: ServicesPageService,
        private serviceDatabasesService: ServiceDatabasesService,
        private serviceRestoresService: ServiceRestoresService,
        private dialogRef: MatDialogRef<RestoreServiceBackupsDialogComponent>,
        @Inject(MAT_DIALOG_DATA) public name: string,
    ) {
        super();
        this.formGroup = this.fb.group({
            database: ['', [Validators.required]],
        });
    }

    ngOnInit(): void {
        this.serviceDatabases$ = this.serviceDatabasesService.getDatabases(
            this.servicesPageService.getCurrentServiceId()
        );
    }

    onSave(): void {
        if (this.checkForm()) {
            this.serviceRestoresService.restore(this.servicesPageService.getCurrentServiceId(),
                this.name, this.formGroup.value.database).subscribe(
                () => {
                    this.messages.success('Database was successfully restored');
                    this.dialogRef.close();
                }
            );
        }
    }

}
