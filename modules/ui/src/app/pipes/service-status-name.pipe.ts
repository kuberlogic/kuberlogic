import { Pipe, PipeTransform } from '@angular/core';
import { ServiceModelStatus } from '@models/service.model';

const statusName = {
    [ServiceModelStatus.RUNNING]: 'Running',
    [ServiceModelStatus.READY]: 'Ready',
    [ServiceModelStatus.FAILED]: 'Failed',
    [ServiceModelStatus.PROCESSING]: 'Processing',
    [ServiceModelStatus.UNKNOWN]: 'Unknown',
    [ServiceModelStatus.NOT_READY]: 'Not Ready',
};

@Pipe({
    name: 'klServiceStatusName'
})
export class ServiceStatusNamePipe implements PipeTransform {

    transform(status: ServiceModelStatus | undefined): string {
        return !!status ? statusName[status] : statusName[ServiceModelStatus.UNKNOWN];
    }

}
