import { ServiceModelStatus } from '@models/service.model';
import { ServiceStatusNamePipe } from './service-status-name.pipe';

describe('ServiceStatusNamePipe', () => {
    const pipe = new ServiceStatusNamePipe();

    it('should create an instance', () => {
        expect(pipe).toBeTruthy();
    });

    it('should return "Running" when status is Running', () => {
        const result = pipe.transform(ServiceModelStatus.RUNNING);
        expect(result).toEqual('Running');
    });

    it('should return "Failed" when status is Failed', () => {
        const result = pipe.transform(ServiceModelStatus.FAILED);
        expect(result).toEqual('Failed');
    });

    it('should return "Processing" when status is Processing', () => {
        const result = pipe.transform(ServiceModelStatus.PROCESSING);
        expect(result).toEqual('Processing');
    });

    it('should return "Unknown" when status is Unknown', () => {
        const result = pipe.transform(ServiceModelStatus.UNKNOWN);
        expect(result).toEqual('Unknown');
    });

    it('should return "Unknown" when status is undefined', () => {
        const result = pipe.transform(undefined);
        expect(result).toEqual('Unknown');
    });
});
