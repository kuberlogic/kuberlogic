import { TestBed } from '@angular/core/testing';

import { MatDialog, MatDialogRef } from '@angular/material/dialog';
import { MockMatDialog } from '@testing/mock-mat-dialog';
import { MockMatDialogRef } from '@testing/mock-mat-dialog-ref';
import { ConfirmDialogService } from './confirm-dialog.service';

describe('ConfirmDialogService', () => {
    let service: ConfirmDialogService;
    let dialog: MockMatDialog;

    beforeEach(() => {
        TestBed.configureTestingModule({
            providers: [
                { provide: MatDialogRef, useClass: MockMatDialogRef },
                { provide: MatDialog, useClass: MockMatDialog },
            ]
        });
        service = TestBed.inject(ConfirmDialogService);
        // @ts-ignore
        dialog = TestBed.inject(MatDialog);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('should open dialog on confirm', () => {
        const spy = spyOn(dialog, 'open').and.callThrough();
        service.confirm('title', 'message');

        expect(spy).toHaveBeenCalled();
    });
});
