import { NavigationExtras } from '@angular/router';
import { Observable, of } from 'rxjs';

export class MockRouter {
    isActiveValue = true;

    navigate(_route: any, _extras?: NavigationExtras): void { // eslint-disable-line
    }

    navigateByUrl(_route: any, _extras?: NavigationExtras): void { // eslint-disable-line
    }

    isActive(_url: string): boolean {
        return this.isActiveValue;
    }

    get events(): Observable<any> {
        return of([]);
    }
}
