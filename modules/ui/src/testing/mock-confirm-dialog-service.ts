import { BehaviorSubject, Observable } from 'rxjs';

export class MockConfirmDialogService {
    resultSub = new BehaviorSubject<boolean>(true);

    confirm(
        _title: string,
        _message: string,
        _buttonConfirmText = 'Yes',
        _buttonCancelText = 'No',
    ): Observable<boolean> {
        return this.resultSub.asObservable();
    }
}
