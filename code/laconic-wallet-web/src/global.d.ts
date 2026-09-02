// Extends the Window interface for Android WebView communication
declare global {
  interface Window {
    // Android bridge callbacks for signature and accounts related events
    Android?: {
      // Called when signature is successfully generated
      onSignatureComplete?: (signature: string) => void;

      // Called when signature generation fails
      onSignatureError?: (error: string) => void;

      // Called when signature process is cancelled
      onSignatureCancelled?: () => void;

      // Called when accounts are ready for use
      onAccountsReady?: () => void;
    };

    // Handles incoming signature requests from Android
    receiveSignRequestFromAndroid?: (message: string) => void;
  }
}

export {};
