import { StyleSheet } from 'react-native';

const styles = StyleSheet.create({
  createWalletContainer: {
    marginTop: 20,
    width: 150,
    alignSelf: 'center',
  },
  signLink: {
    alignItems: 'flex-end',
    marginTop: 24,
  },
  hyperlink: {
    fontWeight: '500',
    textDecorationLine: 'underline',
  },
  highlight: {
    fontWeight: '700',
  },
  accountContainer: {
    padding: 8,
    paddingBottom: 0,
  },
  addAccountButton: {
    marginTop: 24,
    alignSelf: 'center',
  },
  accountComponent: {
    flex: 4,
  },
  appContainer: {
    flexGrow: 1,
    marginTop: 24,
    paddingHorizontal: 24,
  },
  resetContainer: {
    flex: 1,
    justifyContent: 'center',
  },
  resetButton: {
    alignSelf: 'center',
  },
  signButton: {
    marginTop: 20,
    width: 150,
    alignSelf: 'center',
  },
  signPage: {
    paddingHorizontal: 24,
  },
  addNetwork: {
    paddingHorizontal: 24,
    marginTop: 30,
  },
  accountInfo: {
    marginTop: 12,
    marginBottom: 30,
  },
  networkDropdown: {
    marginBottom: 20,
  },
  dialogTitle: {
    padding: 10,
  },
  dialogContents: {
    marginTop: 24,
    padding: 10,
    borderWidth: 1,
    borderRadius: 10,
  },
  dialogWarning: {
    color: 'red',
  },
  gridContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'center',
  },
  gridItem: {
    width: '25%',
    margin: 8,
    padding: 6,
    borderWidth: 1,
    borderColor: '#ccc',
    borderRadius: 8,
    alignItems: 'center',
    justifyContent: 'flex-start',
  },
  HDcontainer: {
    marginTop: 24,
    paddingHorizontal: 8,
  },
  HDrowContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  HDtext: {
    color: 'black',
    fontSize: 18,
    margin: 4,
  },
  HDtextInput: {
    flex: 1,
  },
  HDbuttonContainer: {
    marginTop: 20,
    width: 200,
    alignSelf: 'center',
  },
  spinnerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  LoadingText: {
    color: 'black',
    fontSize: 18,
    padding: 10,
  },
  requestMessage: {
    borderWidth: 1,
    borderRadius: 5,
    marginTop: 50,
    height: 'auto',
    alignItems: 'center',
    justifyContent: 'center',
    padding: 10,
  },
  requestDirectMessage: {
    borderWidth: 1,
    borderRadius: 5,
    marginTop: 20,
    marginBottom: 50,
    height: 500,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 8,
  },
  approveTransfer: {
    height: '40%',
    marginBottom: 30,
  },
  buttonContainer: {
    flexDirection: 'row',
    marginLeft: 20,
    marginBottom: 10,
    justifyContent: 'space-evenly',
  },
  badRequestContainer: {
    alignItems: 'center',
    justifyContent: 'center',
    padding: 20,
  },
  invalidMessageText: {
    color: 'black',
    fontSize: 16,
    textAlign: 'center',
    marginBottom: 20,
  },
  container: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 10,
    paddingHorizontal: 20,
  },
  modalContentContainer: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    borderRadius: 34,
    borderBottomStartRadius: 0,
    borderBottomEndRadius: 0,
    borderWidth: 1,
    width: '100%',
    height: '50%',
    position: 'absolute',
    backgroundColor: 'white',
    bottom: 0,
  },
  modalOuterContainer: { flex: 1 },
  dappLogo: {
    width: 50,
    height: 50,
    borderRadius: 8,
    marginVertical: 16,
    overflow: 'hidden',
  },
  space: {
    width: 50,
  },
  flexRow: {
    display: 'flex',
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: 20,
    paddingHorizontal: 16,
    marginBottom: 10,
  },
  marginVertical8: {
    marginVertical: 8,
    textAlign: 'center',
  },
  subHeading: {
    textAlign: 'center',
    fontWeight: 'bold',
    marginBottom: 10,
    marginTop: 10,
  },
  centerText: {
    textAlign: 'center',
  },
  messageBody: {
    borderWidth: 1,
    borderRadius: 6,
    paddingVertical: 10,
    paddingHorizontal: 10,
    marginVertical: 3,
  },
  cameraContainer: {
    justifyContent: 'center',
    alignItems: 'center',
  },
  inputContainer: {
    marginTop: 20,
  },
  camera: {
    width: 400,
    height: 400,
  },
  dappDetails: {
    display: 'flex',
    alignItems: 'center',
  },
  dataBoxContainer: {
    marginBottom: 10,
  },
  dataBoxLabel: {
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 3,
    color: 'black',
  },
  dataBox: {
    borderWidth: 1,
    borderColor: '#ccc',
    padding: 10,
    borderRadius: 5,
  },
  dataBoxData: {
    fontSize: 16,
    color: 'black',
  },
  transactionText: {
    padding: 8,
    fontSize: 18,
    fontWeight: 'bold',
    color: 'black',
  },
  balancePadding: {
    padding: 8,
  },
  noActiveSessions: { display: 'flex', alignItems: 'center', marginTop: 12 },
  disconnectSession: { display: 'flex', justifyContent: 'center' },
  sessionItem: { paddingLeft: 12, borderBottomWidth: 0.5 },
  sessionsContainer: { paddingLeft: 12, borderBottomWidth: 0.5 },
  walletConnectUriText: { padding: 10 },
  walletConnectLogo: { width: 24, height: 15, margin: 0 },
  selectNetworkText: {
    fontWeight: 'bold',
    marginVertical: 10,
  },
  transactionFeesInput: { marginBottom: 10 },
  approveTransaction: {
    flexGrow: 1,
    marginTop: 0,
    paddingHorizontal: 24,
    paddingVertical: 5,
  },
  transactionLabel: {
    fontWeight: '700',
    padding: 8,
  },
});

export default styles;
