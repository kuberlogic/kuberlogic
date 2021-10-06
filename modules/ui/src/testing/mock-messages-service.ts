import { MessageTypes } from '@services/messages.service';

export class MockMessageService {
    success(_message: string, _action: string = ''): void {}

    error(_message: string, _action: string = ''): void {}

    info(_message: string, _action: string = ''): void {}

    add(_message: string, _action: string, _type: MessageTypes): void {}
}
