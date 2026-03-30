/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.client;

import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import org.rgt.ByteArrayBuffer;
import org.rgt.ServerStatus;
import org.rgt.TerminalException;
import org.rgt.TerminalServerConfiguration;
import org.rgt.TerminalServerListener;
import org.rgt.auth.AuthenticatorException;
import org.rgt.auth.UserAuthenticator;
import java.util.ArrayList;
import java.util.Collection;
import java.util.List;
import org.rgt.TerminalServer;
import org.rgt.admin.AdminOperation;
import static org.rgt.admin.AdminOperation.*;
import org.rgt.admin.AdminResponseCode;
import java.io.IOException;
import java.net.InetAddress;
import java.net.SocketTimeoutException;
import java.net.UnknownHostException;
import java.nio.channels.UnresolvedAddressException;
import java.nio.file.attribute.BasicFileAttributes;
import java.time.Duration;
import java.time.Instant;
import java.util.Arrays;
import java.util.Date;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Iterator;
import java.util.Map;
import java.util.Set;
import org.rgt.Credential;
import org.rgt.CredentialProvider;
import org.rgt.Operation;
import org.rgt.ResponseCode;
import org.rgt.Session;
import org.rgt.TerminalUtil;
import org.rgt.TextScreen;
import org.rgt.auth.TerminalUser;
import org.rgt.auth.UserRepository;
import org.rgt.protocol.ProtocolErrorException;
import org.rgt.protocol.ProtocolProvider;
import org.rgt.protocol.Request;
import org.rgt.protocol.Response;
import org.rgt.protocol.admin.login.LoginRequest;
import org.rgt.protocol.admin.login.LoginResponse;
import org.rgt.protocol.admin.service.ChangeServiceStatusResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.rgt.protocol.Protocol;
import org.rgt.protocol.TransferMonitor;
import org.rgt.protocol.admin.files.GetFileRequest;
import org.rgt.protocol.admin.files.GetFileResponse;
import org.rgt.protocol.admin.files.ListFilesRequest;
import org.rgt.protocol.admin.files.ListFilesResponse;
import org.rgt.protocol.admin.files.PutFileRequest;
import org.rgt.protocol.admin.files.RemoveFileRequest;
import org.rgt.protocol.admin.log.SetLogLevelRequest;
import org.rgt.protocol.admin.server.GetConfigResponse;
import org.rgt.protocol.admin.server.ServerInfoResponse;
import org.rgt.protocol.admin.server.SetConfigRequest;
import org.rgt.protocol.admin.session.GetSessionStatsRequest;
import org.rgt.protocol.admin.session.GetSessionStatsResponse;
import org.rgt.protocol.admin.session.GetSessionsResponse;
import org.rgt.protocol.admin.session.GetTerminalScreenRequest;
import org.rgt.protocol.admin.session.GetTerminalScreenResponse;
import org.rgt.protocol.admin.session.KillAllSessionsResponse;
import org.rgt.protocol.admin.session.KillSessionRequest;
import org.rgt.protocol.admin.users.AddUserRequest;
import org.rgt.protocol.admin.users.GetUsersResponse;
import org.rgt.protocol.admin.users.RemoveUserRequest;
import org.rgt.protocol.admin.users.SetUsersRequest;

/**
 *
 * @author fabio_uggeri
 */
public class RemoteTerminalServer implements TerminalServer {

   private static final Logger LOG = LoggerFactory.getLogger(RemoteTerminalServer.class);

   private static final int RESPONSE_HEADER_SIZE = 6;

   private static final int MINIMUM_ADMIN_PROTOCOL_VERSION = 3;

   private static final int TRANSFER_FILE_CHUNK_SIZE = 512 * 1024;

   private final String id;

   private String remoteServerAddress;

   private final String serverAddress;

   private final int serverPort;

   private final TerminalServerConfiguration configuration = new TerminalServerConfiguration();

   private ServerStatus status = ServerStatus.DISCONNECTED;

   private final List<TerminalServerListener> listeners = new ArrayList<>();

   private UserAuthenticator userAuthenticator = null;

   private UserRepository userRepository;

   private final ProtocolProvider protocolProvider;

   private Connection connection = null;

   private final Map<Long, RemoteSession> sessions = new HashMap<>();

   private int sessionsCount = 0;

   private long startTime = 0;

   private final CredentialProvider credentialProvider;

   private boolean readOnly = false;

   private short protocolVersion = ADMIN_PROTOCOL_VERSION;

   private String serverVersion = "";

   private String userEditing = "";

   private String lastViewedRemotePath = ".";

   public RemoteTerminalServer(String id, String address, int port, ProtocolProvider protocolProvider,
           CredentialProvider credentialProvider) throws AuthenticatorException {
      final RemoteTerminalUserAuthenticator remoteTerminalUserAuthenticator = new RemoteTerminalUserAuthenticator(this);
      this.id = id;
      this.remoteServerAddress = address;
      this.serverAddress = address;
      this.serverPort = port;
      this.userRepository = remoteTerminalUserAuthenticator;
      this.userAuthenticator = remoteTerminalUserAuthenticator;
      this.protocolProvider = protocolProvider;
      this.credentialProvider = credentialProvider;
   }

   private synchronized void writeRequest(final ByteArrayBuffer buffer) throws IOException, TerminalException {
      int write;
      Instant lastWrite = Instant.now();
      while (buffer.remaining() > 0) {
         write = connection.write(buffer.getByteBuffer());
         if (write > 0) {
            lastWrite = Instant.now();
         } else if (write == 0 && Duration.between(lastWrite, Instant.now()).toMillis() >= Connection.DEFAULT_IO_TIMEOUT) {
            throw new TerminalException(AdminResponseCode.SOCKET, "Timeout writing data from server " + serverAddress + ".");
         } else if (write == -1) {
            throw new TerminalException(AdminResponseCode.SOCKET, "Socket channel closed when writing data to server "
                    + serverAddress + ".");
         }
      }
   }

   private synchronized ByteArrayBuffer readResponse() throws TerminalException, IOException {
      int read;
      final ByteArrayBuffer header = new ByteArrayBuffer(RESPONSE_HEADER_SIZE);
      final ByteArrayBuffer body;
      int bodyLen;
      Instant lastRead = Instant.now();
      while (header.remaining() > 0) {
         read = connection.read(header.getByteBuffer());
         if (read > 0) {
            lastRead = Instant.now();
         } else if (read == 0 && Duration.between(lastRead, Instant.now()).toMillis() >= Connection.DEFAULT_IO_TIMEOUT) {
            throw new TerminalException(AdminResponseCode.SOCKET, "Timeout reading data from server " + serverAddress + ".");
         } else if (read == -1) {
            throw new TerminalException(AdminResponseCode.SOCKET, "Socket channel closed when reading header from server "
                    + serverAddress + ".");
         }
      }
      header.flip();
      bodyLen = header.getInt();
      if (bodyLen < 2) {
         LOG.error("Invalid body len in admin message: {}", bodyLen);
         throw new TerminalException(AdminResponseCode.SOCKET, "Invalid body len in admin message from server "
                 + serverAddress + ": " + bodyLen);
      }
      body = new ByteArrayBuffer(bodyLen);
      body.putShort(header.getShort());
      while (body.remaining() > 0) {
         read = connection.read(body.getByteBuffer());
         if (read > 0) {
            lastRead = Instant.now();
         } else if (read == 0 && Duration.between(lastRead, Instant.now()).toMillis() >= Connection.DEFAULT_IO_TIMEOUT) {
            throw new TerminalException(AdminResponseCode.SOCKET, "Timeout reading data from server " + serverAddress + ".");
         } else if (read == -1) {
            throw new TerminalException(AdminResponseCode.SOCKET, "Socket channel closed when reading body from server "
                    + serverAddress + ".");
         }
      }
      body.flip();
      return body;
   }

   private void closeConnection() {
      try {
         if (connection != null) {
            connection.close();
         }
      } catch (IOException ex) {
         LOG.error("Error closing socket connected in server " + serverAddress, ex);
      } finally {
         connection = null;
         sessions.clear();
         sessionsCount = 0;
         serverVersion = "";
         userEditing = "";
         startTime = 0L;
         setStatus(ServerStatus.DISCONNECTED);
      }
   }

   private TerminalException error(final ResponseCode code, final String msg, final Exception ex) throws TerminalException {
      LOG.error(msg, ex);
      if (! (ex instanceof SocketTimeoutException)) {
         closeConnection();
      }
      return new TerminalException(code, msg, ex);
   }

   private TerminalException error(final ResponseCode code, final String msg) throws TerminalException {
      return error(code, msg, true);
   }

   private TerminalException error(final ResponseCode code, final String msg, boolean closeConn) throws TerminalException {
      LOG.error(msg);
      if (closeConn) {
         closeConnection();
      }
      return new TerminalException(code, msg);
   }

   private TerminalException error(final String msg, final TerminalException ex) throws TerminalException {
      return error(msg, ex, true);
   }

   private TerminalException error(final String msg, final TerminalException ex, boolean closeConn) throws TerminalException {
      LOG.error(msg, ex);
      if (closeConn) {
         closeConnection();
      }
      return ex;
   }

   private <P extends Protocol<? extends Request, ? extends Response>> P getProtocol(Operation operation)
           throws ProtocolErrorException {
      return protocolProvider.getProtocol(operation, protocolVersion);
   }

   private <R extends Response> R sendRequest(final AdminOperation operation) throws IOException,
           TerminalException {
      final Protocol<Request, R> protocol = getProtocol(operation);
      final ByteArrayBuffer buffer = new ByteArrayBuffer();
      protocol.putRequest(new Request(operation), buffer);
      writeRequest(buffer);
      return protocol.getResponse(readResponse());
   }

   private <S extends Request, R extends Response> R sendRequest(final S request) throws IOException,
           TerminalException {
      final Protocol<S, R> protocol = getProtocol(request.getOperation());
      final ByteArrayBuffer buffer = new ByteArrayBuffer();
      protocol.putRequest(request, buffer);
      writeRequest(buffer);
      return protocol.getResponse(readResponse());
   }

   public boolean connect() throws TerminalException {
      if (ServerStatus.DISCONNECTED == status) {
         try {
            boolean requestNewCredentials = false;
            for (;;) {
               final LoginResponse response;
               final Credential userCredentials = requestCredential(serverAddress, requestNewCredentials);
               if (userCredentials == null) {
                  setStatus(ServerStatus.DISCONNECTED);
                  return false;
               }
               if (connection != null) {
                  disconnect();
               }
               setStatus(ServerStatus.CONNECTING);
               connection = SocketChannelConnection.open(serverAddress, serverPort);
               response = sendRequest(new LoginRequest(protocolVersion,
                       userCredentials.getUsername(),
                       userCredentials.getPassword()));
               if (response.getResponseCode().value() == AdminResponseCode.SUCCESS.value()) {
                  sessionsCount = response.sessionsCount();
                  startTime = response.startTime();
                  setStatus(response.serverStatus());
                  setReadOnly(response.readOnly());
                  if (response.adminProtocolVersion() > MINIMUM_ADMIN_PROTOCOL_VERSION) {
                     protocolVersion = response.adminProtocolVersion();
                     serverVersion = response.serverVersion();
                     userEditing = response.userEditing();
                  } else {
                     protocolVersion = MINIMUM_ADMIN_PROTOCOL_VERSION;
                  }
                  registerCredential(serverAddress, userCredentials);
                  return true;
               } else if (response.getResponseCode() != AdminResponseCode.INVALID_CREDENTIAL) {
                  setStatus(ServerStatus.DISCONNECTED);
                  throw error(response.getResponseCode(), "Error connecting to server " + serverAddress
                          + ": " + response.getMessage());
               } else if (interactiveCredentialProvider()) {
                  fireNotification(TerminalUtil.getMessage("AdminClientWindow.invalidCredential.msg") + "\n"
                          + TerminalUtil.getMessage("AdminClientWindow.connecting.msg", this));
               } else {
                  setStatus(ServerStatus.DISCONNECTED);
                  fireNotification(TerminalUtil.getMessage("AdminClientWindow.invalidCredential.msg"));
                  return false;
               }
               requestNewCredentials = true;
            }
         } catch (TerminalException ex) {
            setStatus(ServerStatus.DISCONNECTED);
            throw error("Error connecting to server " + serverAddress, ex);
         } catch (UnresolvedAddressException | IOException ex) {
            setStatus(ServerStatus.DISCONNECTED);
            throw error(AdminResponseCode.SOCKET, "Error connecting to server " + serverAddress, ex);
         }
      }
      return false;
   }

   public boolean disconnect() throws TerminalException {
      if (isConnected()) {
         setStatus(ServerStatus.DISCONNECTING);
         try {
            final Response response = sendRequest(LOGOFF);
            if (response.getResponseCode() != AdminResponseCode.SUCCESS) {
               LOG.error("Error disconnecting from remote server {}: {}", serverAddress, response.getMessage());
            }
         } catch (IOException ex) {
            LOG.error("Error disconnecting from remote server " + serverAddress, ex);
         } finally {
            closeConnection();
         }
         return true;
      }
      return false;
   }

   private void setStatus(ServerStatus newStatus) {
      if (newStatus != status) {
         final ServerStatus previouStatus = status;
         status = newStatus;
         fireStatusChanged(previouStatus);
      }
   }

   @Override
   public void startTerminalEmulationService() throws TerminalException {
      if (isConnected()) {
         setRemoteConfiguration();
         startRemoteServer();
         updateLocalStatus();
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
   }

   private boolean startRemoteServer() throws TerminalException {
      try {
         final ChangeServiceStatusResponse response;

         setStatus(ServerStatus.STARTING);
         response = sendRequest(START_SERVICE);
         if (AdminResponseCode.SUCCESS == response.getResponseCode()) {
            setStatus(response.status());
            fireServiceStarted();
            return true;
         } else {
            fireNotification("Error starting service on remote server " + serverAddress + ": " + response.getMessage());
         }
      } catch (IOException ex) {
         throw error(AdminResponseCode.UNKNOWN_ERROR, "Error starting service on remote server " + serverAddress, ex);
      }
      return false;
   }

   @Override
   public void stopTerminalEmulationService() throws TerminalException {
      if (isConnected()) {
         ServerStatus finalStatus = ServerStatus.STOPPED;
         try {
            final ChangeServiceStatusResponse response;
            setStatus(ServerStatus.STOPPING);
            response = sendRequest(STOP_SERVICE);
            if (response.isSuccess()) {
               finalStatus = response.status();
            } else {
               fireNotification("Error stoping service on remote server " + serverAddress + ": " + response.getMessage());
            }
         } catch (IOException ex) {
            throw error(AdminResponseCode.SOCKET, "Error stopping service on server " + serverAddress, ex);
         } catch (TerminalException ex) {
            throw error("Error stopping service on remote server " + serverAddress, ex);
         } finally {
            sessions.clear();
            sessionsCount = 0;
            startTime = 0L;
            setStatus(finalStatus);
            fireServiceStopped();
         }
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
   }

   @Override
   public Collection<RemoteSession> getSessions() throws TerminalException {
      if (isConnected()) {
         updateSessions();
         return sessions.values();
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
   }

   @Override
   public int getSessionsCount() {
      return sessionsCount;
   }
   
   @Override
   public SessionStats getSessionStats(long sessionId) throws TerminalException {
      if (isConnected()) {
         try {
            final GetSessionStatsResponse response = sendRequest(new GetSessionStatsRequest(sessionId));
            if (response.isSuccess()) {
               return response.getSessionStats();
            } else {
               fireNotification("Failed getting stats session from server " + serverAddress + ": " + response.getMessage());
            }
         } catch (IOException ex) {
            throw error(AdminResponseCode.SOCKET, "Error getting stats from session " + id + " on server " + serverAddress);
         } catch (TerminalException ex) {
            throw error("Error getting stats from session " + id + " on server " + serverAddress, ex);
         }
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
      return null;
      
   }

   @Override
   public boolean killSession(long id) throws TerminalException {
      if (isConnected()) {
         try {
            final RemoteSession session = sessions.get(id);
            final Response response = sendRequest(new KillSessionRequest(id));
            if (session != null) {
               sessions.remove(id);
               if (sessionsCount > 0) {
                  --sessionsCount;
               }
               fireSessionClose(session);
            }
            if (response.isSuccess()) {
               return true;
            } else {
               fireNotification("Remote session not closed on server " + serverAddress + ": " + response.getMessage());
            }
         } catch (IOException ex) {
            throw error(AdminResponseCode.SOCKET, "Error killing session " + id + " on server " + serverAddress);
         } catch (TerminalException ex) {
            throw error("Error killing session " + id + " on server " + serverAddress, ex);
         }
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
      return false;
   }

   public int killAllSessions() throws TerminalException {
      if (isConnected()) {
         try {
            final KillAllSessionsResponse response = sendRequest(KILL_ALL_SESSIONS);
            if (response.isSuccess()) {
               sessions.forEach((sessionId, session) -> fireSessionClose(session));
               sessions.clear();
               return response.killedSessions();
            } else {
               fireNotification("Error closing remote sessions on server " + serverAddress + ": " + response.getMessage());
            }
         } catch (IOException ex) {
            throw error(AdminResponseCode.SOCKET, "Error closing sessions on server " + serverAddress);
         } catch (TerminalException ex) {
            throw error("Error closing remote sessions on server " + serverAddress, ex);
         }
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
      return 0;
   }

   @Override
   public Session getSession(long id) {
      return sessions.get(id);
   }

   @Override
   public TextScreen getTerminalScreen(long sessionId) throws TerminalException {
      TerminalException error = null;
      if (isConnected()) {
         try {
            final GetTerminalScreenResponse response = sendRequest(new GetTerminalScreenRequest(sessionId));
            if (response.isSuccess()) {
               return response.screen();
            } else {
               fireNotification("Error getting terminal screen for session " + sessionId + ": " + response.getMessage());
               error = error(AdminResponseCode.SESSION_NOT_FOUND, "Error getting terminal screen for session " + sessionId + ": "
                       + response.getMessage(), false);
            }
         } catch (IOException ex) {
            throw error(AdminResponseCode.SOCKET, "Error getting terminal screen for session " + sessionId
                    + " on server " + serverAddress);
         } catch (TerminalException ex) {
            throw error("Error getting terminal screen for session " + sessionId + " on server " + serverAddress, ex);
         }
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
      throw error;
   }

   @Override
   public ServerStatus getStatus() {
      return status;
   }

   @Override
   public boolean isEmbedded() {
      return false;
   }

   @Override
   public boolean isConnected() {
      return ServerStatus.DISCONNECTED != status
              && ServerStatus.CONNECTING != status
              && connection != null
              && connection.isConnected();
   }

   @Override
   public void addListener(TerminalServerListener listener) {
      listeners.add(listener);
   }

   @Override
   public void removeListener(TerminalServerListener listener) {
      listeners.remove(listener);
   }

   @Override
   public void clearListeners() {
      listeners.clear();
   }

   @Override
   public UserAuthenticator getUserAuthenticator() {
      return userAuthenticator;
   }

   @Override
   public UserRepository getUserRepository() {
      return userRepository;
   }

   @Override
   public void setUserAuthenticator(UserAuthenticator authenticator) {
      this.userAuthenticator = authenticator;
   }

   @Override
   public void setUserRepository(UserRepository repository) {
      this.userRepository = repository;
   }

   @Override
   public ProtocolProvider getProtocolProvider() {
      return protocolProvider;
   }

   public String getRemoteServerAddress() {
      return remoteServerAddress;
   }

   public void setRemoteServerAddress(String remoteServerAddress) {
      this.remoteServerAddress = remoteServerAddress;
   }

   @Override
   public TerminalServerConfiguration getConfiguration() {
      return configuration;
   }

   @Override
   public long getStartTime() {
      return startTime;
   }

   @Override
   public String getId() {
      return id;
   }

   private void fireStateUpdate() {
      listeners.forEach(l -> l.stateUpdate(this));
   }

   private void fireStatusChanged(ServerStatus previousStatus) {
      listeners.forEach(l -> l.statusChanged(this, previousStatus));
   }

   private void fireSessionClose(final Session session) {
      listeners.forEach(l -> l.sessionClose(session));
   }

   private void fireConfigurationSaved() {
      listeners.forEach(l -> l.configurationSaved(this));
   }

   private void fireConfigurationLoaded() {
      listeners.forEach(l -> l.configurationLoaded(this));
   }

   private void fireNotification(final String message) {
      listeners.forEach(l -> l.notification(this, message + "\n"));
   }

   private void fireServiceStarted() {
      listeners.forEach(l -> l.serviceStart(this));
   }

   private void fireServiceStopped() {
      listeners.forEach(l -> l.serviceStop(this));
   }

   @Override
   public boolean saveConfiguration() throws TerminalException {
      if (isConnected()) {
         if (setRemoteConfiguration() && saveRemoteConfiguration()) {
            fireConfigurationSaved();
            return true;
         } else {
            fireNotification("Error saving configuration of server " + serverAddress);
            return false;
         }
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
   }

   private boolean saveRemoteConfiguration() throws TerminalException {
      try {
         final Response response = sendRequest(SAVE_CONFIG);
         if (response.isSuccess()) {
            return true;
         } else {
            fireNotification("Error saving configuration on server " + serverAddress + ": " + response.getMessage());
         }
      } catch (IOException ex) {
         throw error(AdminResponseCode.SOCKET, "Error saving configuration on server " + serverAddress, ex);
      } catch (TerminalException ex) {
         throw error("Error saving configuration on server " + serverAddress + ".", ex);
      }
      return false;
   }

   @Override
   public void loadConfiguration() throws TerminalException {
      if (isConnected()) {
         loadRemoteConfiguration();
         getRemoteConfiguration();
         fireConfigurationLoaded();
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
   }

   private boolean loadRemoteConfiguration() throws TerminalException {
      try {
         final Response response = sendRequest(LOAD_CONFIG);
         if (response.isSuccess()) {
            return true;
         } else {
            fireNotification("Error loading configuration on server " + serverAddress + ": " + response.getMessage());
         }
      } catch (IOException ex) {
         throw error(AdminResponseCode.SOCKET, "Error loading configuration on server " + serverAddress, ex);
      } catch (TerminalException ex) {
         throw error("Error loading configuration on server " + serverAddress + ".", ex);
      }
      return false;
   }

   private void updateSessions(final Set<RemoteSession> remoteSessions) {
      Iterator<RemoteSession> it = sessions.values().iterator();
      while (it.hasNext()) {
         final RemoteSession s = it.next();
         if (!remoteSessions.contains(s)) {
            fireSessionClose(s);
            it.remove();
         }
      }
      remoteSessions.forEach(s -> {
         final RemoteSession localSession = sessions.get(s.getId());
         if (localSession != null) {
            localSession.setTerminalAddress(s.getTerminalAddress());
            localSession.setTerminalUser(s.getTerminalUser());
            localSession.setAppPid(s.getAppPid());
            localSession.setStatus(s.getStatus());
            localSession.setStartTime(s.getStartTime());
            localSession.setCommandLine(s.getCommandLine());
         } else {
            sessions.put(s.getId(), s);
         }
      });
   }

   public boolean updateSessions() throws TerminalException {
      if (isConnected()) {
         try {
            final GetSessionsResponse response = sendRequest(GET_SESSIONS);
            if (response.isSuccess()) {
               updateSessions(new HashSet<>(response.getSessions()));
               sessionsCount = sessions.size();
               return true;
            } else {
               fireNotification("Error getting sessions from server " + serverAddress + ": " + response.getMessage());
            }
         } catch (IOException ex) {
            throw error(AdminResponseCode.SOCKET, "Error getting sessions from server " + serverAddress, ex);
         } catch (TerminalException ex) {
            throw error("Error getting sessions from server " + serverAddress, ex);
         }
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
      return false;
   }

   @Override
   public void updateLocalStatus() throws TerminalException {
      if (isConnected()) {
         getRemoteStatus();
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
   }

   @Override
   public void updateLocalState() throws TerminalException {
      if (isConnected()) {
         getRemoteStatus();
         getRemoteConfiguration();
         updateTerminalUsers();
         fireStateUpdate();
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
   }

   @Override
   public void updateRemoteConfiguration() throws TerminalException {
      if (isConnected()) {
         setRemoteConfiguration();
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
   }

   private boolean getRemoteStatus() throws TerminalException {
      try {
         final ServerInfoResponse response = sendRequest(GET_STATUS);
         if (response.isSuccess()) {
            setStatus(response.status());
            sessionsCount = response.sessionsCount();
            startTime = response.startTime();
            return true;
         } else {
            fireNotification("Error updating status from server " + serverAddress + ": " + response.getMessage());
         }
      } catch (IOException ex) {
         throw error(AdminResponseCode.SOCKET, "Error updating status from server " + serverAddress, ex);
      } catch (TerminalException ex) {
         throw error("Error updating status from server " + serverAddress + ".", ex);
      }
      return false;
   }

   public boolean setRemoteConfiguration() throws TerminalException {
      if (isConnected()) {
         try {
            final Response response = sendRequest(new SetConfigRequest(configuration.stringValues(), true));
            if (response.isSuccess()) {
               return true;
            } else {
               fireNotification("Error setting configuration on server " + serverAddress + ": " + response.getMessage());
            }
         } catch (IOException ex) {
            throw error(AdminResponseCode.SOCKET, "Error setting configuration on server " + serverAddress, ex);
         } catch (TerminalException ex) {
            throw error("Error setting configuration on server " + serverAddress + ".", ex);
         }
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
      return false;
   }

   private boolean getRemoteConfiguration() throws TerminalException {
      if (isConnected()) {
         try {
            final GetConfigResponse response = sendRequest(GET_CONFIG);
            if (response.isSuccess()) {
               configuration.setValues(response.config());
               return true;
            } else {
               fireNotification("Error getting configuration from server " + serverAddress + ": " + response.getMessage());
            }
         } catch (IOException ex) {
            throw error(AdminResponseCode.SOCKET, "Error getting configuration from server " + serverAddress, ex);
         } catch (TerminalException ex) {
            throw error("Error getting configuration from server " + serverAddress, ex);
         }
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
      return false;
   }

   @Override
   public void updateLogLevel() throws TerminalException {
      setRemoteLogLevel();
   }

   public boolean setRemoteLogLevel() throws TerminalException {
      if (isConnected()) {
         try {
            final SetLogLevelRequest request = new SetLogLevelRequest(configuration.appLogLevel().value(),
                    configuration.serverLogLevel().value(),
                    configuration.teLogLevel().value());
            final Response response = sendRequest(request);
            if (response.isSuccess()) {
               return true;
            } else {
               fireNotification("Error setting log level on server " + serverAddress + ": " + response.getMessage());
            }
         } catch (IOException ex) {
            throw error(AdminResponseCode.SOCKET, "Error setting log level on server " + serverAddress, ex);
         } catch (TerminalException ex) {
            throw error("Error setting log level on server " + serverAddress, ex);
         }
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
      return false;
   }

   public boolean setRemoteTerminalUsers() throws TerminalException {
      if (isConnected() && userRepository != null) {
         try {
            final SetUsersRequest request = new SetUsersRequest(userRepository.getUsers());
            final Response response = sendRequest(request);
            if (response.isSuccess()) {
               return true;
            } else {
               fireNotification("Error sending user to server " + serverAddress + ": " + response.getMessage());
            }
         } catch (IOException ex) {
            throw error(AdminResponseCode.SOCKET, "Error sending users to server " + serverAddress, ex);
         } catch (TerminalException ex) {
            throw error("Error sending user to server " + serverAddress, ex);
         }
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
      return false;
   }

   public boolean updateTerminalUsers() throws TerminalException {
      if (isConnected()) {
         try {
            final GetUsersResponse response = sendRequest(GET_USERS);
            if (response.isSuccess()) {
               if (response.hasUserAuthentication() && userRepository != null) {
                  userRepository.clearUsers();
                  for (TerminalUser u : response.getUsers()) {
                     userRepository.addUser(u);
                  }
               } else {
                  this.userRepository = null;
               }
               return true;
            } else {
               fireNotification("Error getting users from remote server " + serverAddress + ": " + response.getMessage());
            }
         } catch (IOException ex) {
            throw error(AdminResponseCode.SOCKET, "Error getting users from server " + serverAddress, ex);
         } catch (TerminalException ex) {
            throw error("Error getting users from remote server " + serverAddress, ex);
         }
      } else {
         throw error(AdminResponseCode.CONNECTION_LOST, "Connection lost with server " + serverAddress);
      }
      return false;
   }

   public boolean saveRemoteTerminalUsers() throws TerminalException {
      try {
         final Response response = sendRequest(SAVE_USERS);
         if (response.isSuccess()) {
            return true;
         } else {
            fireNotification("Error saving users on server " + serverAddress + ": " + response.getMessage());
         }
      } catch (IOException ex) {
         throw error(AdminResponseCode.UNKNOWN_ERROR, "Error saving users on server " + serverAddress, ex);
      }
      return false;
   }

   public boolean loadRemoteTerminalUsers() throws TerminalException {
      try {
         final Response response = sendRequest(LOAD_USERS);
         if (response.isSuccess()) {
            return true;
         } else {
            fireNotification("Error loading users on server " + serverAddress + ": " + response.getMessage());
         }
      } catch (IOException ex) {
         throw error(AdminResponseCode.UNKNOWN_ERROR, "Error loading users on server " + serverAddress, ex);
      }
      return false;
   }

   public boolean addRemoteTerminalUser(final TerminalUser user) throws TerminalException {
      try {
         final Response response = sendRequest(new AddUserRequest(user));
         if (response.isSuccess()) {
            return true;
         } else {
            fireNotification("Error adding user to server " + serverAddress + ": " + response.getMessage());
         }
      } catch (IOException ex) {
         throw error(AdminResponseCode.UNKNOWN_ERROR, "Error adding user to server " + serverAddress, ex);
      }
      return false;
   }

   public boolean removeRemoteTerminalUser(final String username) throws TerminalException {
      try {
         final Response response = sendRequest(new RemoveUserRequest(username));
         if (response.isSuccess()) {
            return true;
         } else {
            fireNotification("Error removing user from server " + serverAddress + ": " + response.getMessage());
         }
      } catch (IOException ex) {
         throw error(AdminResponseCode.UNKNOWN_ERROR, "Error removing user from server " + serverAddress, ex);
      }
      return false;
   }

   @Override
   public void startAdminService() throws TerminalException {
   }

   @Override
   public void stopAdminService() throws TerminalException {
   }

   @Override
   public boolean isRunning() {
      if (isConnected()) {
         return ServerStatus.RUNNING == status;
      }
      return false;
   }

   @Override
   public boolean isRunningAdminService() {
      return true;
   }

   @Override
   public String toString() {
      return serverAddress;
   }

   @Override
   public InetAddress getInetAddress() {
      try {
         return InetAddress.getByName(remoteServerAddress);
      } catch (UnknownHostException ex) {
         return null;
      }
   }

   @Override
   public void lostConnection(Session session) {
   }

   public boolean killAdminSessions() throws TerminalException {
      if (connection == null) {
         try (final SocketChannelConnection killConnection = SocketChannelConnection.open(serverAddress, serverPort)) {
            final Response response;
            connection = killConnection;
            response = sendRequest(KILL_ADMIN_SESSIONS);
            if (response.isSuccess()) {
               return true;
            } else {
               fireNotification("Error killing admin sessions from server " + serverAddress + ": " + response.getMessage());
            }
         } catch (IOException ex) {
            throw error(AdminResponseCode.UNKNOWN_ERROR, "Error killing admin sessions from server " + serverAddress, ex);
         } finally {
            connection = null;
         }
      }
      return false;
   }

   public RemoteFolder listRemoteFiles(final String path) throws TerminalException {
      try {
         final ListFilesResponse response = sendRequest(new ListFilesRequest(path));
         if (response.isSuccess()) {
            return new RemoteFolder(response.folderPathname()).files(response.filesInfo());
         } else {
            fireNotification("Error listing remote files from path '" + path + "' on server " + serverAddress + ": "
                    + response.getMessage());
         }
      } catch (IOException ex) {
         throw error(AdminResponseCode.UNKNOWN_ERROR, "Error listing remote files from path '" + path + "' on server "
                 + serverAddress, ex);
      }
      return new RemoteFolder(path);
   }

   public boolean downloadFile(final String remotePathname, final File localFile, boolean force, TransferMonitor tm)
           throws TerminalException {
      try {
         GetFileRequest request;
         GetFileResponse response;
         if (localFile.exists() && !force) {
            fireNotification("Local file already exists:  " + localFile);
            return false;
         }
         request = new GetFileRequest(remotePathname);
         response = sendRequest(request);
         if (!response.isSuccess()) {
            fireNotification("Error downloading remote file '" + remotePathname + "' from server " + serverAddress
                    + ": " + response.getMessage());
            return false;
         }
         try (final FileOutputStream fos = new FileOutputStream(localFile, false)) {
            long fileLen = response.fileInfo().getLength();
            TerminalUtil.setCreationTime(localFile, response.fileInfo().getCreationTime());
            TerminalUtil.setLastModificationTime(localFile, response.fileInfo().getLastModificationTime());
            request.pathFilename("");
            while (fileLen > 0) {
               fos.write(response.data());
               if (tm != null && !tm.transfer(response.dataSize())) {
                  sendRequest(AdminOperation.CANCEL);
                  fireNotification("Operation canceled by user");
                  return false;
               }
               fileLen -= response.dataSize();
               if (fileLen > 0) {
                  response = sendRequest(request);
                  if (!response.isSuccess()) {
                     fireNotification("Error downloading remote file '" + remotePathname + "' from server " + serverAddress
                             + ": " + response.getMessage());
                     return false;
                  }
               }
            }
         }
      } catch (IOException ex) {
         throw error(AdminResponseCode.UNKNOWN_ERROR, "Error downloading remote file '" + remotePathname + "' from server "
                 + serverAddress, ex);
      }
      return true;
   }

   public boolean uploadFile(final File localFile, final String remotePathname, boolean force, TransferMonitor tm)
           throws TerminalException {
      try {
         final BasicFileAttributes fileAttr = TerminalUtil.getAttributes(localFile);
         final PutFileRequest request;
         Response response;
         if (!localFile.exists()) {
            fireNotification("Local file does not exist:  " + localFile);
            return false;
         } else if (!localFile.isFile()) {
            fireNotification(localFile + " is not a file");
            return false;
         }
         request = new PutFileRequest(new File(remotePathname, localFile.getName()).getPath())
                 .force(force)
                 .fileSize(localFile.length())
                 .creationTime(new Date(fileAttr.creationTime().toMillis()))
                 .lastModificationTime(new Date(fileAttr.creationTime().toMillis()));
         try (final FileInputStream fis = new FileInputStream(localFile)) {
            final byte[] data = new byte[TRANSFER_FILE_CHUNK_SIZE];
            int bytesRead;
            do {
               bytesRead = fis.read(data);
               if (bytesRead < data.length) {
                  request.data(Arrays.copyOf(data, bytesRead));
               } else {
                  request.data(data);
               }
               response = sendRequest(request);
               if (tm != null && !tm.transfer(bytesRead)) {
                  sendRequest(AdminOperation.CANCEL);
                  fireNotification("Operation canceled by user");
                  return false;
               }
               request.filePathname("");
            } while (response.isSuccess() && bytesRead == data.length);
            if (!response.isSuccess()) {
               fireNotification("Error uploading file '" + localFile + "' to '" + remotePathname + "' on server " + serverAddress
                       + ": " + response.getMessage());
               return false;
            }
         }
      } catch (IOException ex) {
         throw error(AdminResponseCode.UNKNOWN_ERROR, "Error downloading remote file '" + remotePathname + "' from server "
                 + serverAddress, ex);
      }
      return true;
   }

   public boolean removeFile(final String remotePathname, final String filename) throws TerminalException {
      try {
         RemoveFileRequest request;
         Response response;
         request = new RemoveFileRequest(remotePathname, filename);
         response = sendRequest(request);
         if (!response.isSuccess()) {
            fireNotification("Error removing remote file '" + filename + "' at '" + remotePathname
                    + "' from server " + serverAddress + ": " + response.getMessage());
            return false;
         }
      } catch (IOException ex) {
         throw error(AdminResponseCode.UNKNOWN_ERROR, "Error removing remote file '" + filename + "' at '" + remotePathname
                 + "' from server " + serverAddress, ex);
      }
      return true;
   }

   private Credential requestCredential(final String serverAddress, final boolean requestNew) {
      return credentialProvider.getCredential(serverAddress, requestNew);
   }

   private void registerCredential(final String serverAddress, final Credential credential) {
      credentialProvider.registerCredential(serverAddress, credential);
   }
   
   private boolean interactiveCredentialProvider() {
      return credentialProvider.isInteractive();
   }

   @Override
   public boolean isReadOnly() {
      return this.readOnly;
   }

   public void setReadOnly(boolean readOnly) {
      this.readOnly = readOnly;
   }

   @Override
   public String getVersion() {
      return serverVersion;
   }

   public String getUserEditing() {
      return userEditing;
   }

   public String getLastViewedRemotePath() {
      return lastViewedRemotePath;
   }

   public void setLastViewedRemotePath(String lastViewedRemotePath) {
      this.lastViewedRemotePath = lastViewedRemotePath;
   }
}
