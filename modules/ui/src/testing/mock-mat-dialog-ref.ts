import { Subject } from 'rxjs';

export class MockMatDialogRef {
    afterClosedSub = new Subject<any>();
    afterClosed = () => this.afterClosedSub;

    // eslint-disable-next-line
    close(value: any): void {
        this.afterClosedSub.next(value);
    }
}
