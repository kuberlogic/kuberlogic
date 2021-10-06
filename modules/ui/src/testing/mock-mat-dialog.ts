import { MatDialogRef } from '@angular/material/dialog';
import { Subject } from 'rxjs';

export class MockMatDialog {
    closeSource = new Subject<boolean>();
    openResult = <MatDialogRef<any>> {
        afterClosed: () => {
            return this.closeSource.asObservable();
        },
        // eslint-disable-next-line
        close: () => {
        },
    };

    open(): MatDialogRef<any> {
        return this.openResult;
    }
}
