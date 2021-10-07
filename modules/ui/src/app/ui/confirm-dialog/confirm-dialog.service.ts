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

import { Injectable } from '@angular/core';
import { MatDialog, MatDialogRef } from '@angular/material/dialog';
import { ConfirmDialogComponent } from '@ui/confirm-dialog/confirm-dialog.component';
import { Observable } from 'rxjs';

@Injectable({
    providedIn: 'root'
})
export class ConfirmDialogService {

    private dialogRef!: MatDialogRef<ConfirmDialogComponent>;

    constructor(private dialog: MatDialog) {}

    confirm(
        title: string,
        message: string,
        buttonConfirmText = 'Yes',
        buttonCancelText = 'No',
    ): Observable<boolean> {
        const dialogData = {
            title,
            message,
            buttonConfirmText,
            buttonCancelText
        };
        this.dialogRef = this.dialog.open(ConfirmDialogComponent, {
            width: '400px',
            data: dialogData
        });

        return this.dialogRef.afterClosed();
    }
}
