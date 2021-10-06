import { TextOnlySnackBar } from '@angular/material/snack-bar/simple-snack-bar';
import { MatSnackBarConfig } from '@angular/material/snack-bar/snack-bar-config';
import { MatSnackBarRef } from '@angular/material/snack-bar/snack-bar-ref';

export class MockMatSnackbar {
    open(_message: string, _action?: string, _config?: MatSnackBarConfig): MatSnackBarRef<TextOnlySnackBar> {
        return {} as MatSnackBarRef<TextOnlySnackBar>;
    }

    dismiss(): void {}
}
