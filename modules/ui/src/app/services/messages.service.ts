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
