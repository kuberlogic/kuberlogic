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
import { MatSnackBar } from '@angular/material/snack-bar';

export type MessageTypes = 'error' | 'info' | 'success';

/**
 * This class shows snackBar with message
 * color of snackBar depends on type
 */
@Injectable({
    providedIn: 'root'
})
export class MessagesService {
    constructor(private snackBar: MatSnackBar) {}

    success(message: string, action: string = ''): void {
        this.add(message, action, 'success');
    }

    error(message: string, action: string = 'x'): void {
        this.add(message, action, 'error');
    }

    info(message: string, action: string = ''): void {
        this.add(message, action, 'info');
    }

    /**
     * Opens snackBar
     * @param message text
     * @param action button text
     * @param type ['info', 'error', 'success']
     */
    add(message: string, action: string, type: MessageTypes): void {
        this.snackBar.open(message, action, {
            duration: type === 'error' ? 0 : 5000,
            verticalPosition: 'bottom',
            horizontalPosition: 'right',
            panelClass: [`mat-snack-bar-container--${type}`]
        });
    }
}
