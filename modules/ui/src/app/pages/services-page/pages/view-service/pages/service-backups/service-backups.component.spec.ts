/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';

import { HttpClientTestingModule } from '@angular/common/http/testing';
import { MatDialog } from '@angular/material/dialog';
import { RestoreServiceBackupsDialogComponent } from '@pages/services-page/pages/view-service/pages/service-backups/components/restore-service-backups-dialog/restore-service-backups-dialog.component';
import { MessagesService } from '@services/messages.service';
import { ServiceBackupsService } from '@services/service-backups.service';
import { MockMatDialog } from '@testing/mock-mat-dialog';
import { MockMessageService } from '@testing/mock-messages-service';
import { MockServiceBackupsService } from '@testing/mock-service-backups-service';
import { ServiceBackupsComponent } from './service-backups.component';

describe('ServiceBackupsComponent', () => {
    let component: ServiceBackupsComponent;
    let fixture: ComponentFixture<ServiceBackupsComponent>;
    let backupsService: MockServiceBackupsService;
    let dialog: MockMatDialog;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ServiceBackupsComponent],
            schemas: [NO_ERRORS_SCHEMA],
            imports: [HttpClientTestingModule],
            providers: [
                { provide: ServiceBackupsService, useClass: MockServiceBackupsService },
                { provide: MatDialog, useClass: MockMatDialog },
                { provide: MessagesService, useClass: MockMessageService },
                RestoreServiceBackupsDialogComponent,
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ServiceBackupsComponent);
        component = fixture.componentInstance;
        // @ts-ignore
        backupsService = TestBed.inject(ServiceBackupsService);
        // @ts-ignore
        dialog = TestBed.inject(MatDialog);
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should open "RestoreServiceBackupsDialogComponent" dialog on restore', () => {
        const spy = spyOn(dialog, 'open').and.callThrough();
        component.onRestore('db1');
        fixture.detectChanges();

        // @ts-ignore
        expect(spy.calls.mostRecent().args[0]).toEqual(RestoreServiceBackupsDialogComponent);
    });
});
