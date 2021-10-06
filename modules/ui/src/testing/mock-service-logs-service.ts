import { ServiceLogModel } from '@models/service-log.model';
import { Observable, of } from 'rxjs';

export class MockServiceLogsService {
    logs: ServiceLogModel[] =  [
        {
            lines: 100,
            body: 'sgwertert',
        },
        {
            lines: 100,
            body: 'sdfgdsfhgfdg',
        }
    ];

    get(_serviceId: string): Observable<ServiceLogModel[] | undefined> {
        return of(this.logs);
    }
}
