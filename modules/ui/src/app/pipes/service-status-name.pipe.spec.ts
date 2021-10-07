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
