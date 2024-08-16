export interface ShareInput {
  documentId: number;
  invitedEmails?: string[];
  sharedByEmail?: string;
  expirationDate?: Date;
  passphrase?: string;
  shareToken: string; // 8-character random string
  accessType: 'anyone-with-link' | 'organization' | 'passphrase-protected' | 'only-invited';
}