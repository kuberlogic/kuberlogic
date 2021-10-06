export interface ServiceBackupConfigModel {
    enabled: boolean;
    aws_access_key_id: string;
    aws_secret_access_key: string;
    bucket: string;
    endpoint: string;
    schedule: string;
}
