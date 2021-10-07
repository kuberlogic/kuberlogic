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

import { TestBed } from '@angular/core/testing';

import { MatSnackBar } from '@angular/material/snack-bar';
import { MockMatSnackbar } from '@testing/mock-mat-snackbar';
import { MessagesService } from './messages.service';

describe('MessagesService', () => {
    let service: MessagesService;

    beforeEach(() => {
        TestBed.configureTestingModule({
            providers: [
                { provide: MatSnackBar, useClass: MockMatSnackbar }
            ]
        });
        service = TestBed.inject(MessagesService);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('should add success type message with action', () => {
        const spy = spyOn(service, 'add').and.callThrough();
        service.success('message', 'action');
        expect(spy).toHaveBeenCalledWith('message', 'action', 'success');
    });

    it('should add success type message without action', () => {
        const spy = spyOn(service, 'add').and.callThrough();
        service.success('message');
        expect(spy).toHaveBeenCalledWith('message', '', 'success');
    });

    it('should add error type message with action', () => {
        const spy = spyOn(service, 'add').and.callThrough();
        service.error('message', 'action');
        expect(spy).toHaveBeenCalledWith('message', 'action', 'error');
    });

    it('should add error type message without action', () => {
        const spy = spyOn(service, 'add').and.callThrough();
        service.error('message');
        expect(spy).toHaveBeenCalledWith('message', 'x', 'error');
    });

    it('should add info type message with action', () => {
        const spy = spyOn(service, 'add').and.callThrough();
        service.info('message', 'action');
        expect(spy).toHaveBeenCalledWith('message', 'action', 'info');
    });

    it('should add info type message without action', () => {
        const spy = spyOn(service, 'add').and.callThrough();
        service.info('message');
        expect(spy).toHaveBeenCalledWith('message', '', 'info');
    });
});
