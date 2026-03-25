/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.gui;

import org.rgt.protocol.admin.files.RemoteFileOperationListener;
import org.rgt.PropertiesUtil;
import org.rgt.ServerStatus;
import org.rgt.Session;
import org.rgt.TerminalException;
import org.rgt.TerminalServerListener;
import org.rgt.auth.AuthenticatorException;
import org.rgt.client.RemoteTerminalServer;
import org.rgt.gui.table.ServerTableModel;
import java.awt.Dimension;
import java.awt.HeadlessException;
import java.awt.Toolkit;
import java.awt.event.KeyEvent;
import java.io.File;
import java.io.IOException;
import java.util.List;
import javax.swing.JOptionPane;
import javax.swing.UIManager;
import javax.swing.event.ListSelectionEvent;
import javax.swing.event.ListSelectionListener;
import javax.swing.table.TableColumnModel;
import org.ini4j.Profile.Section;
import org.ini4j.Wini;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.rgt.TerminalServer;
import java.awt.event.WindowEvent;
import java.awt.event.WindowListener;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;
import java.util.Properties;
import javax.swing.ImageIcon;
import javax.swing.JFileChooser;
import javax.swing.JLabel;
import javax.swing.SwingConstants;
import javax.swing.event.TableModelEvent;
import javax.swing.filechooser.FileNameExtensionFilter;
import javax.swing.text.BadLocationException;
import org.rgt.Credential;
import org.rgt.CredentialProvider;
import org.rgt.Notifier;
import org.rgt.SerialExecutor;
import org.rgt.TerminalUtil;
import org.rgt.client.Security;
import org.rgt.gui.table.DecoratorTableCellRenderer;
import org.rgt.gui.table.TableCellRendererDecorator;
import org.rgt.gui.table.TableFilter;
import org.rgt.protocol.DefaultProtocolProvider;
import org.rgt.protocol.ProtocolProvider;
import org.rgt.protocol.admin.files.FileInfo;

/**
 *
 * @author fabio_uggeri
 */
public class MainWindow
        extends javax.swing.JFrame
        implements TerminalServerListener, WindowListener, CredentialProvider, ListSelectionListener, RemoteFileOperationListener,
        Notifier {

   private static final long serialVersionUID = 9177369344836885134L;

   private static final String DEFAULT_SERVERS_LIST_NAME = "servers.ini";

   private static final int SERVER_ID = 0;
   private static final int SERVER_ADDRESS = 1;
   private static final int SERVER_STATUS = 2;
   private static final int SERVER_SESSIONS = 3;
   private static final int SERVER_STARTUP = 4;
   private static final int SERVER_VERSION = 5;

   private static final Logger LOG = LoggerFactory.getLogger(MainWindow.class);

   private final int MESSAGE_AREA_MAX_LEN = 32 * 1024;

   private final File PREFERENCES_FILE = new File(System.getProperty("user.home", "."), "rgt_preferences.properties");

   private final Map<TerminalServer, ServerDetailsWindow> monitorWindows = new HashMap<>();
   
   private final Map<TerminalServer, SessionsWindow> sessionsWindows = new HashMap<>();

   private final SerialExecutor executor = SerialExecutor.newInstance("MainExecutor");

   private File serversFile = null;

   private final ProtocolProvider protocolProvider = new DefaultProtocolProvider();

   private final HashMap<String, Credential> credentials = new HashMap<>();

   private Credential lastCredential = null;

   private Credential lastTriedCredential = null;

   private boolean batchOperation = false;

   /**
    * Creates new form ServersWindow
    */
   public MainWindow() {
      initComponents();
      loadPreferences();
      configureServersTable();
      loadConfigurations();
   }

   private void configureServersTable() {
      final TableColumnModel colModel = serversTable.getColumnModel();
      final ImageIcon readonlyIcon = new ImageIcon(getClass().getResource("/readonly16.png"));
      final DecoratorTableCellRenderer statusDecorator;

      statusDecorator = new DecoratorTableCellRenderer((TableCellRendererDecorator<ServerTableModel, JLabel>) (table, model, label, row, col, sel, focus) -> {
         final RemoteTerminalServer server = ((ServerTableModel) model).getServerAt(row);
         if (server.isConnected() && server.isReadOnly()) {
            label.setIcon(readonlyIcon);
            label.setHorizontalTextPosition(SwingConstants.LEFT);
         } else {
            label.setIcon(null);
            label.setHorizontalTextPosition(SwingConstants.TRAILING);
         }
      });

      serversTable.setModel(new ServerTableModel());
      serversTable.getTableHeader().setToolTipText(TerminalUtil.getMessage("AdminClientWindow.serversTable.header.tooltip"));
      colModel.getColumn(SERVER_ID).setResizable(true);
      colModel.getColumn(SERVER_ADDRESS).setResizable(true);
      colModel.getColumn(SERVER_STATUS).setResizable(true);
      colModel.getColumn(SERVER_STATUS).setCellRenderer(statusDecorator);
      colModel.getColumn(SERVER_SESSIONS).setResizable(true);
      colModel.getColumn(SERVER_STARTUP).setResizable(true);
      colModel.getColumn(SERVER_VERSION).setResizable(true);
      colModel.getColumn(SERVER_ID).setPreferredWidth(100);
      colModel.getColumn(SERVER_ADDRESS).setPreferredWidth(140);
      colModel.getColumn(SERVER_STATUS).setPreferredWidth(50);
      colModel.getColumn(SERVER_SESSIONS).setPreferredWidth(30);
      colModel.getColumn(SERVER_STARTUP).setPreferredWidth(90);
      colModel.getColumn(SERVER_VERSION).setPreferredWidth(50);
      serversTable.setColumnSelectionAllowed(false);
      serversTable.setRowSelectionAllowed(true);
      serversTable.getSelectionModel().addListSelectionListener(this);
      TableFilter.forTable(serversTable).apply().repaint();
   }

   /**
    * This method is called from within the constructor to initialize the form.
    * WARNING: Do NOT modify this code. The content of this method is always
    * regenerated by the Form Editor.
    */
   @SuppressWarnings("unchecked")
   // <editor-fold defaultstate="collapsed" desc="Generated Code">//GEN-BEGIN:initComponents
   private void initComponents() {

      jPanel1 = new javax.swing.JPanel();
      jToolBar1 = new javax.swing.JToolBar();
      btnOpen = new javax.swing.JButton();
      btnSaveAs = new javax.swing.JButton();
      btnSave = new javax.swing.JButton();
      jSeparator3 = new javax.swing.JToolBar.Separator();
      btnAdd = new javax.swing.JButton();
      btnRemove = new javax.swing.JButton();
      btnEdit = new javax.swing.JButton();
      jSeparator2 = new javax.swing.JToolBar.Separator();
      btnRefresh = new javax.swing.JButton();
      jSeparator4 = new javax.swing.JToolBar.Separator();
      btnServer = new javax.swing.JButton();
      btnConnect = new javax.swing.JButton();
      btnDisconnect = new javax.swing.JButton();
      jSeparator6 = new javax.swing.JToolBar.Separator();
      btnKillAdminSessions = new javax.swing.JButton();
      jSeparator1 = new javax.swing.JToolBar.Separator();
      btnStopService = new javax.swing.JButton();
      btnStartService = new javax.swing.JButton();
      btnRestartService = new javax.swing.JButton();
      btnConfigServers = new javax.swing.JButton();
      jSeparator10 = new javax.swing.JToolBar.Separator();
      btnKillSessions = new javax.swing.JButton();
      btnSessions = new javax.swing.JButton();
      jSeparator9 = new javax.swing.JToolBar.Separator();
      btnRemoteFiles = new javax.swing.JButton();
      jToolBar2 = new javax.swing.JToolBar();
      btnClose = new javax.swing.JButton();
      splitPanel = new javax.swing.JSplitPane();
      jScrollPane1 = new javax.swing.JScrollPane();
      serversTable = new javax.swing.JTable();
      jPanel2 = new javax.swing.JPanel();
      jScrollPane2 = new javax.swing.JScrollPane();
      messageArea = new javax.swing.JTextArea();
      jToolBar3 = new javax.swing.JToolBar();
      btnCopyMessages = new javax.swing.JButton();
      btnCutMessages = new javax.swing.JButton();
      jSeparator5 = new javax.swing.JToolBar.Separator();
      btnClearMessages = new javax.swing.JButton();
      jPanel3 = new javax.swing.JPanel();
      activeConnectionsLabel = new javax.swing.JLabel();
      activeConnections = new javax.swing.JTextField();
      sessionsCountLabel = new javax.swing.JLabel();
      totalSessions = new javax.swing.JTextField();

      setDefaultCloseOperation(javax.swing.WindowConstants.EXIT_ON_CLOSE);
      java.util.ResourceBundle bundle = java.util.ResourceBundle.getBundle("org/rgt/gui/Bundle"); // NOI18N
      setTitle(bundle.getString("MainWindow.title")); // NOI18N
      setIconImage(new javax.swing.ImageIcon(getClass().getResource("/server.png")).getImage());
      setName("Form"); // NOI18N
      addWindowListener(new java.awt.event.WindowAdapter() {
         public void windowClosed(java.awt.event.WindowEvent evt) {
            formWindowClosed(evt);
         }
      });

      jPanel1.setName("jPanel1"); // NOI18N

      jToolBar1.setBorder(javax.swing.BorderFactory.createEmptyBorder(1, 1, 1, 1));
      jToolBar1.setRollover(true);
      jToolBar1.setName("jToolBar1"); // NOI18N

      btnOpen.setIcon(new javax.swing.ImageIcon(getClass().getResource("/open_folder_32.png"))); // NOI18N
      btnOpen.setToolTipText(bundle.getString("MainWindow.btnOpen.toolTipText")); // NOI18N
      btnOpen.setFocusable(false);
      btnOpen.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnOpen.setName("btnOpen"); // NOI18N
      btnOpen.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnOpen.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnOpenActionPerformed(evt);
         }
      });
      jToolBar1.add(btnOpen);

      btnSaveAs.setIcon(new javax.swing.ImageIcon(getClass().getResource("/save_as_32.png"))); // NOI18N
      btnSaveAs.setToolTipText(bundle.getString("MainWindow.btnSaveAs.toolTipText")); // NOI18N
      btnSaveAs.setFocusable(false);
      btnSaveAs.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnSaveAs.setName("btnSaveAs"); // NOI18N
      btnSaveAs.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnSaveAs.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnSaveAsActionPerformed(evt);
         }
      });
      jToolBar1.add(btnSaveAs);

      btnSave.setIcon(new javax.swing.ImageIcon(getClass().getResource("/diskette.png"))); // NOI18N
      btnSave.setToolTipText(bundle.getString("MainWindow.btnSave.toolTipText")); // NOI18N
      btnSave.setEnabled(false);
      btnSave.setFocusable(false);
      btnSave.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnSave.setName("btnSave"); // NOI18N
      btnSave.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnSave.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnSaveActionPerformed(evt);
         }
      });
      jToolBar1.add(btnSave);

      jSeparator3.setName("jSeparator3"); // NOI18N
      jToolBar1.add(jSeparator3);

      btnAdd.setIcon(new javax.swing.ImageIcon(getClass().getResource("/add.png"))); // NOI18N
      btnAdd.setToolTipText(bundle.getString("MainWindow.btnAdd.toolTipText")); // NOI18N
      btnAdd.setFocusable(false);
      btnAdd.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnAdd.setName("btnAdd"); // NOI18N
      btnAdd.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnAdd.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnAddActionPerformed(evt);
         }
      });
      jToolBar1.add(btnAdd);

      btnRemove.setIcon(new javax.swing.ImageIcon(getClass().getResource("/remove.png"))); // NOI18N
      btnRemove.setToolTipText(bundle.getString("MainWindow.btnRemove.toolTipText")); // NOI18N
      btnRemove.setEnabled(false);
      btnRemove.setFocusable(false);
      btnRemove.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnRemove.setName("btnRemove"); // NOI18N
      btnRemove.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnRemove.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnRemoveActionPerformed(evt);
         }
      });
      jToolBar1.add(btnRemove);

      btnEdit.setIcon(new javax.swing.ImageIcon(getClass().getResource("/pencil.png"))); // NOI18N
      btnEdit.setToolTipText(bundle.getString("MainWindow.btnEdit.toolTipText")); // NOI18N
      btnEdit.setEnabled(false);
      btnEdit.setFocusable(false);
      btnEdit.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnEdit.setName("btnEdit"); // NOI18N
      btnEdit.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnEdit.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnEditActionPerformed(evt);
         }
      });
      jToolBar1.add(btnEdit);

      jSeparator2.setName("jSeparator2"); // NOI18N
      jToolBar1.add(jSeparator2);

      btnRefresh.setIcon(new javax.swing.ImageIcon(getClass().getResource("/refresh.png"))); // NOI18N
      btnRefresh.setToolTipText(bundle.getString("MainWindow.btnRefresh.toolTipText")); // NOI18N
      btnRefresh.setEnabled(false);
      btnRefresh.setFocusable(false);
      btnRefresh.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnRefresh.setName("btnRefresh"); // NOI18N
      btnRefresh.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnRefresh.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnRefreshActionPerformed(evt);
         }
      });
      jToolBar1.add(btnRefresh);

      jSeparator4.setName("jSeparator4"); // NOI18N
      jToolBar1.add(jSeparator4);

      btnServer.setIcon(new javax.swing.ImageIcon(getClass().getResource("/computer.png"))); // NOI18N
      btnServer.setToolTipText(bundle.getString("MainWindow.btnServer.toolTipText")); // NOI18N
      btnServer.setEnabled(false);
      btnServer.setFocusable(false);
      btnServer.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnServer.setName("btnServer"); // NOI18N
      btnServer.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnServer.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnServerActionPerformed(evt);
         }
      });
      jToolBar1.add(btnServer);

      btnConnect.setIcon(new javax.swing.ImageIcon(getClass().getResource("/connect.png"))); // NOI18N
      btnConnect.setToolTipText(bundle.getString("MainWindow.btnConnect.toolTipText")); // NOI18N
      btnConnect.setEnabled(false);
      btnConnect.setFocusable(false);
      btnConnect.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnConnect.setName("btnConnect"); // NOI18N
      btnConnect.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnConnect.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnConnectActionPerformed(evt);
         }
      });
      jToolBar1.add(btnConnect);

      btnDisconnect.setIcon(new javax.swing.ImageIcon(getClass().getResource("/disconnect.png"))); // NOI18N
      btnDisconnect.setToolTipText(bundle.getString("MainWindow.btnDisconnect.toolTipText")); // NOI18N
      btnDisconnect.setEnabled(false);
      btnDisconnect.setFocusable(false);
      btnDisconnect.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnDisconnect.setName("btnDisconnect"); // NOI18N
      btnDisconnect.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnDisconnect.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnDisconnectActionPerformed(evt);
         }
      });
      jToolBar1.add(btnDisconnect);

      jSeparator6.setName("jSeparator6"); // NOI18N
      jToolBar1.add(jSeparator6);

      btnKillAdminSessions.setIcon(new javax.swing.ImageIcon(getClass().getResource("/cut_wire.png"))); // NOI18N
      btnKillAdminSessions.setToolTipText(bundle.getString("MainWindow.btnKillAdminSessions.toolTipText")); // NOI18N
      btnKillAdminSessions.setEnabled(false);
      btnKillAdminSessions.setFocusable(false);
      btnKillAdminSessions.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnKillAdminSessions.setName("btnKillAdminSessions"); // NOI18N
      btnKillAdminSessions.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnKillAdminSessions.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnKillAdminSessionsActionPerformed(evt);
         }
      });
      jToolBar1.add(btnKillAdminSessions);

      jSeparator1.setName("jSeparator1"); // NOI18N
      jToolBar1.add(jSeparator1);

      btnStopService.setIcon(new javax.swing.ImageIcon(getClass().getResource("/stop.png"))); // NOI18N
      btnStopService.setToolTipText(bundle.getString("MainWindow.btnStopService.toolTipText")); // NOI18N
      btnStopService.setEnabled(false);
      btnStopService.setFocusable(false);
      btnStopService.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnStopService.setName("btnStopService"); // NOI18N
      btnStopService.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnStopService.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnStopServiceActionPerformed(evt);
         }
      });
      jToolBar1.add(btnStopService);

      btnStartService.setIcon(new javax.swing.ImageIcon(getClass().getResource("/start.png"))); // NOI18N
      btnStartService.setToolTipText(bundle.getString("MainWindow.btnStartService.toolTipText")); // NOI18N
      btnStartService.setEnabled(false);
      btnStartService.setFocusable(false);
      btnStartService.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnStartService.setName("btnStartService"); // NOI18N
      btnStartService.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnStartService.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnStartServiceActionPerformed(evt);
         }
      });
      jToolBar1.add(btnStartService);

      btnRestartService.setIcon(new javax.swing.ImageIcon(getClass().getResource("/restart.png"))); // NOI18N
      btnRestartService.setToolTipText(bundle.getString("MainWindow.btnRestartService.toolTipText")); // NOI18N
      btnRestartService.setEnabled(false);
      btnRestartService.setFocusable(false);
      btnRestartService.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnRestartService.setName("btnRestartService"); // NOI18N
      btnRestartService.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnRestartService.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnRestartServiceActionPerformed(evt);
         }
      });
      jToolBar1.add(btnRestartService);

      btnConfigServers.setIcon(new javax.swing.ImageIcon(getClass().getResource("/config.png"))); // NOI18N
      btnConfigServers.setToolTipText(bundle.getString("MainWindow.btnConfigServers.toolTipText")); // NOI18N
      btnConfigServers.setEnabled(false);
      btnConfigServers.setFocusable(false);
      btnConfigServers.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnConfigServers.setName("btnConfigServers"); // NOI18N
      btnConfigServers.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnConfigServers.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnConfigServersActionPerformed(evt);
         }
      });
      jToolBar1.add(btnConfigServers);

      jSeparator10.setName("jSeparator10"); // NOI18N
      jToolBar1.add(jSeparator10);

      btnKillSessions.setIcon(new javax.swing.ImageIcon(getClass().getResource("/kill_sessions.png"))); // NOI18N
      btnKillSessions.setToolTipText(bundle.getString("MainWindow.btnKillSessions.toolTipText")); // NOI18N
      btnKillSessions.setEnabled(false);
      btnKillSessions.setFocusable(false);
      btnKillSessions.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnKillSessions.setName("btnKillSessions"); // NOI18N
      btnKillSessions.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnKillSessions.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnKillSessionsActionPerformed(evt);
         }
      });
      jToolBar1.add(btnKillSessions);

      btnSessions.setIcon(new javax.swing.ImageIcon(getClass().getResource("/sessions.png"))); // NOI18N
      btnSessions.setToolTipText(bundle.getString("MainWindow.btnSessions.toolTipText")); // NOI18N
      btnSessions.setDefaultCapable(false);
      btnSessions.setEnabled(false);
      btnSessions.setFocusable(false);
      btnSessions.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnSessions.setName("btnSessions"); // NOI18N
      btnSessions.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnSessions.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnSessionsActionPerformed(evt);
         }
      });
      jToolBar1.add(btnSessions);

      jSeparator9.setName("jSeparator9"); // NOI18N
      jToolBar1.add(jSeparator9);

      btnRemoteFiles.setIcon(new javax.swing.ImageIcon(getClass().getResource("/remote_folder_32.png"))); // NOI18N
      btnRemoteFiles.setToolTipText(bundle.getString("MainWindow.btnRemoteFiles.toolTipText")); // NOI18N
      btnRemoteFiles.setEnabled(false);
      btnRemoteFiles.setFocusable(false);
      btnRemoteFiles.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnRemoteFiles.setName("btnRemoteFiles"); // NOI18N
      btnRemoteFiles.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnRemoteFiles.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnRemoteFilesActionPerformed(evt);
         }
      });
      jToolBar1.add(btnRemoteFiles);

      jToolBar2.setBorder(javax.swing.BorderFactory.createEmptyBorder(1, 1, 1, 1));
      jToolBar2.setRollover(true);
      jToolBar2.setName("jToolBar2"); // NOI18N

      btnClose.setIcon(new javax.swing.ImageIcon(getClass().getResource("/exit.png"))); // NOI18N
      btnClose.setToolTipText(bundle.getString("MainWindow.btnClose.toolTipText")); // NOI18N
      btnClose.setFocusable(false);
      btnClose.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnClose.setName("btnClose"); // NOI18N
      btnClose.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnClose.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnCloseActionPerformed(evt);
         }
      });
      jToolBar2.add(btnClose);

      javax.swing.GroupLayout jPanel1Layout = new javax.swing.GroupLayout(jPanel1);
      jPanel1.setLayout(jPanel1Layout);
      jPanel1Layout.setHorizontalGroup(
         jPanel1Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(jPanel1Layout.createSequentialGroup()
            .addComponent(jToolBar1, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.RELATED, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
            .addComponent(jToolBar2, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE))
      );
      jPanel1Layout.setVerticalGroup(
         jPanel1Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addComponent(jToolBar1, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
         .addComponent(jToolBar2, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
      );

      splitPanel.setDividerLocation(250);
      splitPanel.setOrientation(javax.swing.JSplitPane.VERTICAL_SPLIT);
      splitPanel.setCursor(new java.awt.Cursor(java.awt.Cursor.DEFAULT_CURSOR));
      splitPanel.setName("splitPanel"); // NOI18N
      splitPanel.setOneTouchExpandable(true);

      jScrollPane1.setBorder(javax.swing.BorderFactory.createTitledBorder(bundle.getString("MainWindow.jScrollPane1.border.title"))); // NOI18N
      jScrollPane1.setName("jScrollPane1"); // NOI18N

      serversTable.setModel(new ServerTableModel());
      serversTable.setToolTipText("");
      serversTable.setAutoResizeMode(javax.swing.JTable.AUTO_RESIZE_NEXT_COLUMN);
      serversTable.setName("serversTable"); // NOI18N
      serversTable.getTableHeader().setReorderingAllowed(false);
      serversTable.addMouseListener(new java.awt.event.MouseAdapter() {
         public void mouseClicked(java.awt.event.MouseEvent evt) {
            serversTableMouseClicked(evt);
         }
      });
      serversTable.addKeyListener(new java.awt.event.KeyAdapter() {
         public void keyPressed(java.awt.event.KeyEvent evt) {
            serversTableKeyPressed(evt);
         }
      });
      jScrollPane1.setViewportView(serversTable);

      splitPanel.setTopComponent(jScrollPane1);

      jPanel2.setBorder(javax.swing.BorderFactory.createTitledBorder(bundle.getString("MainWindow.jPanel2.border.title"))); // NOI18N
      jPanel2.setName("jPanel2"); // NOI18N

      jScrollPane2.setBorder(null);
      jScrollPane2.setName("jScrollPane2"); // NOI18N

      messageArea.setColumns(80);
      messageArea.setFont(new java.awt.Font("Courier New", 0, 12)); // NOI18N
      messageArea.setRows(4);
      messageArea.setTabSize(3);
      messageArea.setBorder(null);
      messageArea.setName("messageArea"); // NOI18N
      messageArea.addKeyListener(new java.awt.event.KeyAdapter() {
         public void keyPressed(java.awt.event.KeyEvent evt) {
            messageAreaKeyPressed(evt);
         }
         public void keyTyped(java.awt.event.KeyEvent evt) {
            messageAreaKeyTyped(evt);
         }
      });
      jScrollPane2.setViewportView(messageArea);

      jToolBar3.setBorder(javax.swing.BorderFactory.createEmptyBorder(1, 1, 1, 1));
      jToolBar3.setOrientation(javax.swing.SwingConstants.VERTICAL);
      jToolBar3.setRollover(true);
      jToolBar3.setName("jToolBar3"); // NOI18N

      btnCopyMessages.setIcon(new javax.swing.ImageIcon(getClass().getResource("/copy_16.png"))); // NOI18N
      btnCopyMessages.setToolTipText(bundle.getString("MainWindow.btnCopyMessages.toolTipText")); // NOI18N
      btnCopyMessages.setFocusable(false);
      btnCopyMessages.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnCopyMessages.setName("btnCopyMessages"); // NOI18N
      btnCopyMessages.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnCopyMessages.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnCopyMessagesActionPerformed(evt);
         }
      });
      jToolBar3.add(btnCopyMessages);

      btnCutMessages.setIcon(new javax.swing.ImageIcon(getClass().getResource("/cut_16.png"))); // NOI18N
      btnCutMessages.setToolTipText(bundle.getString("MainWindow.btnCutMessages.toolTipText")); // NOI18N
      btnCutMessages.setFocusable(false);
      btnCutMessages.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnCutMessages.setName("btnCutMessages"); // NOI18N
      btnCutMessages.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnCutMessages.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnCutMessagesActionPerformed(evt);
         }
      });
      jToolBar3.add(btnCutMessages);

      jSeparator5.setName("jSeparator5"); // NOI18N
      jToolBar3.add(jSeparator5);

      btnClearMessages.setIcon(new javax.swing.ImageIcon(getClass().getResource("/erase_16.png"))); // NOI18N
      btnClearMessages.setToolTipText(bundle.getString("MainWindow.btnClearMessages.toolTipText")); // NOI18N
      btnClearMessages.setFocusable(false);
      btnClearMessages.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnClearMessages.setName("btnClearMessages"); // NOI18N
      btnClearMessages.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnClearMessages.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnClearMessagesActionPerformed(evt);
         }
      });
      jToolBar3.add(btnClearMessages);

      javax.swing.GroupLayout jPanel2Layout = new javax.swing.GroupLayout(jPanel2);
      jPanel2.setLayout(jPanel2Layout);
      jPanel2Layout.setHorizontalGroup(
         jPanel2Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(jPanel2Layout.createSequentialGroup()
            .addComponent(jToolBar3, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE)
            .addGap(0, 1053, Short.MAX_VALUE))
         .addGroup(jPanel2Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
            .addGroup(javax.swing.GroupLayout.Alignment.TRAILING, jPanel2Layout.createSequentialGroup()
               .addGap(32, 32, 32)
               .addComponent(jScrollPane2, javax.swing.GroupLayout.DEFAULT_SIZE, 1049, Short.MAX_VALUE)))
      );
      jPanel2Layout.setVerticalGroup(
         jPanel2Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addComponent(jToolBar3, javax.swing.GroupLayout.Alignment.TRAILING, javax.swing.GroupLayout.DEFAULT_SIZE, 261, Short.MAX_VALUE)
         .addGroup(jPanel2Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
            .addComponent(jScrollPane2))
      );

      splitPanel.setBottomComponent(jPanel2);

      jPanel3.setBorder(javax.swing.BorderFactory.createEtchedBorder());
      jPanel3.setName("jPanel3"); // NOI18N
      jPanel3.setPreferredSize(new java.awt.Dimension(38, 24));

      activeConnectionsLabel.setLabelFor(activeConnections);
      activeConnectionsLabel.setText(bundle.getString("MainWindow.activeConnectionsLabel.text")); // NOI18N
      activeConnectionsLabel.setName("activeConnectionsLabel"); // NOI18N

      activeConnections.setEditable(false);
      activeConnections.setHorizontalAlignment(javax.swing.JTextField.LEFT);
      activeConnections.setText("0");
      activeConnections.setBorder(null);
      activeConnections.setName("activeConnections"); // NOI18N

      sessionsCountLabel.setLabelFor(totalSessions);
      sessionsCountLabel.setText(bundle.getString("MainWindow.sessionsCountLabel.text")); // NOI18N
      sessionsCountLabel.setName("sessionsCountLabel"); // NOI18N

      totalSessions.setEditable(false);
      totalSessions.setHorizontalAlignment(javax.swing.JTextField.LEFT);
      totalSessions.setText("0");
      totalSessions.setBorder(null);
      totalSessions.setName("totalSessions"); // NOI18N

      javax.swing.GroupLayout jPanel3Layout = new javax.swing.GroupLayout(jPanel3);
      jPanel3.setLayout(jPanel3Layout);
      jPanel3Layout.setHorizontalGroup(
         jPanel3Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(jPanel3Layout.createSequentialGroup()
            .addComponent(activeConnectionsLabel, javax.swing.GroupLayout.PREFERRED_SIZE, 111, javax.swing.GroupLayout.PREFERRED_SIZE)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.RELATED)
            .addComponent(activeConnections, javax.swing.GroupLayout.PREFERRED_SIZE, 91, javax.swing.GroupLayout.PREFERRED_SIZE)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.RELATED)
            .addComponent(sessionsCountLabel, javax.swing.GroupLayout.PREFERRED_SIZE, 53, javax.swing.GroupLayout.PREFERRED_SIZE)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.RELATED)
            .addComponent(totalSessions, javax.swing.GroupLayout.PREFERRED_SIZE, 92, javax.swing.GroupLayout.PREFERRED_SIZE)
            .addGap(0, 0, Short.MAX_VALUE))
      );
      jPanel3Layout.setVerticalGroup(
         jPanel3Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(javax.swing.GroupLayout.Alignment.TRAILING, jPanel3Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.BASELINE)
            .addComponent(activeConnectionsLabel, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
            .addComponent(activeConnections)
            .addComponent(sessionsCountLabel, javax.swing.GroupLayout.PREFERRED_SIZE, 20, javax.swing.GroupLayout.PREFERRED_SIZE)
            .addComponent(totalSessions, javax.swing.GroupLayout.PREFERRED_SIZE, 20, javax.swing.GroupLayout.PREFERRED_SIZE))
      );

      javax.swing.GroupLayout layout = new javax.swing.GroupLayout(getContentPane());
      getContentPane().setLayout(layout);
      layout.setHorizontalGroup(
         layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addComponent(jPanel1, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
         .addComponent(splitPanel, javax.swing.GroupLayout.DEFAULT_SIZE, 1091, Short.MAX_VALUE)
         .addComponent(jPanel3, javax.swing.GroupLayout.DEFAULT_SIZE, 1091, Short.MAX_VALUE)
      );
      layout.setVerticalGroup(
         layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(layout.createSequentialGroup()
            .addComponent(jPanel1, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.RELATED)
            .addComponent(splitPanel, javax.swing.GroupLayout.DEFAULT_SIZE, 539, Short.MAX_VALUE)
            .addGap(0, 0, 0)
            .addComponent(jPanel3, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE))
      );

      pack();
   }// </editor-fold>//GEN-END:initComponents

   private void btnCloseActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnCloseActionPerformed
      exit();
   }//GEN-LAST:event_btnCloseActionPerformed

   private void exit() throws HeadlessException {
      if (btnSave.isEnabled()
              && JOptionPane.showConfirmDialog(this, TerminalUtil.getMessage("AdminClientWindow.save_exit.msg"),
                      TerminalUtil.getMessage("AdminClientWindow.save_config.msg"), JOptionPane.YES_NO_OPTION, JOptionPane.QUESTION_MESSAGE) == 0) {
         saveConfiguration();
      }
      closeAllConnections();
      savePreferences();
      executor.stop(true);
      setVisible(false);
      dispose();
      System.exit(0);
   }

   private void btnAddActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnAddActionPerformed
      addServer();
   }//GEN-LAST:event_btnAddActionPerformed

   private void addServer() throws HeadlessException {
      RemoteServerDetails detailsWin = new RemoteServerDetails(this, true);
      detailsWin.setVisible(true);
      if (detailsWin.isConfirm()) {
         try {
            final RemoteTerminalServer server = new RemoteTerminalServer(detailsWin.getId(), detailsWin.getAddress(),
                    detailsWin.getPort(), protocolProvider, this);
            final ServerTableModel model = (ServerTableModel) serversTable.getModel();
            if (model.indexOfServer(server.getId()) >= 0) {
               message("Server already exists with thid ID\n");
               JOptionPane.showMessageDialog(this, TerminalUtil.getMessage("AdminClientWindow.server_exists.msg"),
                       TerminalUtil.getMessage("AdminClientWindow.server_registered.msg"),
                       JOptionPane.INFORMATION_MESSAGE);
            } else {
               message(TerminalUtil.getMessage("AdminClientWindow.server_added.msg", server), "\n");
               model.addServer(server);
               server.addListener(this);
               btnSave.setEnabled(true);
            }
         } catch (AuthenticatorException ex) {
            LOG.error("Erro ao criar servidor", ex);
            message(TerminalUtil.getMessage("AdminClientWindow.error.msg"), "\n",
                    ex.getMessage(),
                    "\n================================================================================\n");
            JOptionPane.showMessageDialog(this, ex.getLocalizedMessage(),
                    TerminalUtil.getMessage("AdminClientWindow.error_create_server.msg"), JOptionPane.ERROR_MESSAGE);
         }
      }
      detailsWin.dispose();
   }

   private ServerTableModel serversModel() {
      return (ServerTableModel) serversTable.getModel();
   }

   private int[] modelSelectedRows() {
      final int selectedRows[] = serversTable.getSelectedRows();
      for (int i = 0; i < selectedRows.length; i++) {
         selectedRows[i] = serversTable.convertRowIndexToModel(selectedRows[i]);
      }
      return selectedRows;
   }

   private int modelSelectedRow() {
      return serversTable.convertRowIndexToModel(serversTable.getSelectedRow());
   }

   private void btnRemoveActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnRemoveActionPerformed
      if (serversTable.getSelectedRow() >= 0) {
         final int[] selectedRows = modelSelectedRows();
         for (int row : selectedRows) {
            serversModel().removeServer(row);
         }
         btnSave.setEnabled(true);
      }
   }//GEN-LAST:event_btnRemoveActionPerformed

   private void btnEditActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnEditActionPerformed
      if (serversTable.getSelectedRowCount() == 1) {
         editServer(serversTable.convertRowIndexToModel(serversTable.getSelectedRow()));
      }
   }//GEN-LAST:event_btnEditActionPerformed

   private void editServer(final int rowModel) {
      final ServerTableModel model = serversModel();
      final RemoteTerminalServer server = model.getServerAt(rowModel);
      if (server != null) {
         final RemoteServerDetails detailsWin = new RemoteServerDetails(this, true);
         detailsWin.disableIdField();
         detailsWin.setId(server.getId());
         detailsWin.setAddress(server.getRemoteServerAddress());
         detailsWin.setPort(server.getConfiguration().adminPort().value());
         detailsWin.setVisible(true);
         if (detailsWin.isConfirm()) {
            server.setRemoteServerAddress(detailsWin.getAddress());
            server.getConfiguration().adminPort().value(detailsWin.getPort());
            model.notifyServerUpdate(rowModel);
            btnSave.setEnabled(true);
         }
         detailsWin.dispose();
      }
   }

   private void btnSaveActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnSaveActionPerformed
      if (JOptionPane.showConfirmDialog(null, TerminalUtil.getMessage("AdminClientWindow.confirm_save.msg"),
              TerminalUtil.getMessage("AdminClientWindow.save_config.msg"),
              JOptionPane.YES_NO_OPTION, JOptionPane.QUESTION_MESSAGE) == 0) {
         saveConfiguration();
      }
   }//GEN-LAST:event_btnSaveActionPerformed

   private void saveConfiguration() throws HeadlessException {
      try {
         final List<RemoteTerminalServer> servers = ((ServerTableModel) serversTable.getModel()).getServers();
         final Wini ini = new Wini();
         servers.forEach(server -> {
            ini.put(server.getId(), "address", server.getRemoteServerAddress());
            ini.put(server.getId(), "adminPort", server.getConfiguration().adminPort().value());
         });
         ini.store(serversFile);
         message(TerminalUtil.getMessage("AdminClientWindow.config_saved.msg", serversFile), '\n');
         btnSave.setEnabled(false);
         setTitle(TerminalUtil.getMessage("AdminClientWindow.servers_title.msg", serversFile.getName()) + localVersion());
      } catch (IOException ex) {
         message(TerminalUtil.getMessage("AdminClientWindow.error_saving_file.msg", serversFile, '\n'));
         JOptionPane.showMessageDialog(this, ex.getLocalizedMessage(),
                 TerminalUtil.getMessage("AdminClientWindow.error_saving.msg"), JOptionPane.ERROR_MESSAGE);
         LOG.error("Error saving configuration file.", ex);
      }
   }

   private void btnConnectActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnConnectActionPerformed
      final int[] selectedRows = modelSelectedRows();
      for (int row : selectedRows) {
         connectToServer(row);
      }
      executor.execute(() -> verifySelection(modelSelectedRows()));
   }//GEN-LAST:event_btnConnectActionPerformed

   private void connectToServer(final int row) throws HeadlessException {
      final ServerTableModel model = serversModel();
      final RemoteTerminalServer server = model.getServerAt(row);
      if (server != null && !server.isConnected()) {
         executor.execute(() -> message(TerminalUtil.getMessage("AdminClientWindow.connecting.msg", server)));
         executor.execute(() -> {
            try {
               if (server.connect()) {
                  server.updateLocalState();
                  if (server.isReadOnly()) {
                     message(TerminalUtil.getMessage("AdminClientWindow.connect_readonly.msg", server.getUserEditing()), "\n");
                  } else {
                     message(TerminalUtil.getMessage("AdminClientWindow.done.msg"), "\n");
                  }
               }
            } catch (Throwable ex) {
               message(TerminalUtil.getMessage("AdminClientWindow.error.msg"), "\n", ex.getMessage(),
                       "\n================================================================================\n");
               LOG.error("Error connecting to server " + server.getConfiguration().address() + ":"
                       + server.getConfiguration().adminPort(), ex);
            } finally {
               model.notifyServerUpdate(row);
            }
         });
      }
   }

   private void btnDisconnectActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnDisconnectActionPerformed
      final int[] selectedRows = modelSelectedRows();
      for (int row : selectedRows) {
         disconnectFromServer(row);
      }
      executor.execute(() -> verifySelection(modelSelectedRows()));
   }//GEN-LAST:event_btnDisconnectActionPerformed

   private void disconnectFromServer(int row) {
      final ServerTableModel model = serversModel();
      final RemoteTerminalServer server = model.getServerAt(row);
      if (server != null && server.isConnected()) {
         executor.execute(() -> message(TerminalUtil.getMessage("AdminClientWindow.disconnecting.msg", server)));
         executor.execute(() -> {
            try {
               server.disconnect();
               hideServerWindow(server);
            } catch (Throwable ex) {
               LOG.error("Error disconnecting from server " + server.getConfiguration().address() + ":"
                       + server.getConfiguration().adminPort(), ex);
            } finally {
               message(TerminalUtil.getMessage("AdminClientWindow.done.msg"), "\n");
               model.notifyServerUpdate(row);
            }
         });
      }
   }

   private void serversTableMouseClicked(java.awt.event.MouseEvent evt) {//GEN-FIRST:event_serversTableMouseClicked
      final int row = serversTable.rowAtPoint(evt.getPoint());
      if (evt.getClickCount() == 2 && row >= 0) {
         final int modelRow = serversTable.convertRowIndexToModel(row);
         connectToServer(modelRow);
         showServerWindow(modelRow);
         executor.execute(() -> verifySelection(modelSelectedRows()));
      }
   }//GEN-LAST:event_serversTableMouseClicked

   private void btnServerActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnServerActionPerformed
      final int[] selectedRows = modelSelectedRows();
      for (int row : selectedRows) {
         showServerWindow(row);
      }
   }//GEN-LAST:event_btnServerActionPerformed

   private void serversTableKeyPressed(java.awt.event.KeyEvent evt) {//GEN-FIRST:event_serversTableKeyPressed
      if (serversTable.getSelectedRowCount() == 1) {
         final ServerTableModel model = serversModel();
         final int rowModel = serversTable.convertRowIndexToModel(serversTable.getSelectedRow());
         switch (evt.getKeyCode()) {
            case KeyEvent.VK_ENTER:
               connectToServer(rowModel);
               showServerWindow(rowModel);
               break;
            case KeyEvent.VK_DELETE:
               disconnectFromServer(rowModel);
               model.removeServer(rowModel);
               btnSave.setEnabled(true);
               break;
            case KeyEvent.VK_INSERT:
               addServer();
               break;
            default:
               break;
         }
      }
   }//GEN-LAST:event_serversTableKeyPressed

   private void btnRefreshActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnRefreshActionPerformed
      refreshSelectedRows();
   }//GEN-LAST:event_btnRefreshActionPerformed

   private void refreshSelectedRows() {
      final int[] selectedRows = modelSelectedRows();
      for (int row : selectedRows) {
         refreshServerData(row);
      }
      executor.execute(() -> verifySelection(modelSelectedRows()));
   }

   private void refreshServerData(int row) {
      final ServerTableModel model = serversModel();
      final RemoteTerminalServer server = model.getServerAt(row);
      if (server != null) {
         if (server.isConnected()) {
            executor.execute(() -> message(TerminalUtil.getMessage("AdminClientWindow.refreshing.msg", server)));
            executor.execute(() -> {
               try {
                  batchOperation = true;
                  server.updateLocalState();
                  message(TerminalUtil.getMessage("AdminClientWindow.done.msg"), "\n");
               } catch (TerminalException ex) {
                  message(TerminalUtil.getMessage("AdminClientWindow.error.msg"), "\n", ex.getMessage(),
                          "\n================================================================================\n");
                  LOG.error("Error updating status from server " + server.getConfiguration().address(), ex);
                  hideSessionsWindow(server);
               } finally {
                  updateConnectionsAndSessionsCount();
                  batchOperation = false;
                  model.notifyServerUpdate(row);
               }
            });
         } else {
            hideServerWindow(server);
            hideSessionsWindow(server);
            model.notifyServerUpdate(row);
         }
      }
   }

   private void formWindowClosed(java.awt.event.WindowEvent evt) {//GEN-FIRST:event_formWindowClosed
      exit();
   }//GEN-LAST:event_formWindowClosed

   private void removeInitialTextFromMessageArea(int newTextLen) {
      if (messageArea.getDocument().getLength() + newTextLen > MESSAGE_AREA_MAX_LEN) {
         int line = 0;
         try {
            while (line < messageArea.getLineCount()) {
               final int endLineOffset = messageArea.getLineEndOffset(line++);
               if (endLineOffset >= messageArea.getDocument().getLength()) {
                  messageArea.setText("");
                  return;
               } else if (endLineOffset >= newTextLen) {
                  messageArea.getDocument().remove(0, endLineOffset);
                  return;
               }
            }
         } catch (BadLocationException ex) {
            LOG.error("Error cleaning message area", ex);
         }
      }
   }

   @Override
   public void message(Object... values) {
      for (Object value : values) {
         if (value != null) {
            final String text = value.toString();
            removeInitialTextFromMessageArea(text.length());
            messageArea.append(text);
         } else {
            removeInitialTextFromMessageArea(4);
            messageArea.append("null");
         }
      }
      messageArea.setCaretPosition(messageArea.getDocument().getLength());
   }

   private void btnKillAdminSessionsActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnKillAdminSessionsActionPerformed
      final int[] selectedRows = modelSelectedRows();
      for (int row : selectedRows) {
         killServerAdminSessions(row);
      }
      executor.execute(() -> verifySelection(modelSelectedRows()));
   }//GEN-LAST:event_btnKillAdminSessionsActionPerformed

   private void killServerAdminSessions(int rowModel) {
      final ServerTableModel model = serversModel();
      final RemoteTerminalServer server = model.getServerAt(rowModel);
      if (!server.isConnected()) {
         executor.execute(() -> message(TerminalUtil.getMessage("AdminClientWindow.killing_admin.msg", server)));
         executor.execute(() -> {
            try {
               server.killAdminSessions();
               message(TerminalUtil.getMessage("AdminClientWindow.done.msg"), "\n");
            } catch (TerminalException | RuntimeException ex) {
               message(TerminalUtil.getMessage("AdminClientWindow.error.msg"), "\n", ex.getMessage(),
                       "\n================================================================================\n");
               LOG.error("Error killing admin sessions from server "
                       + server.getConfiguration().address() + ":" + server.getConfiguration().adminPort(), ex);
            } finally {
               model.notifyServerUpdate(rowModel);
            }
         });
      }
   }

   private void btnCopyMessagesActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnCopyMessagesActionPerformed
      copyMessageArea();
   }//GEN-LAST:event_btnCopyMessagesActionPerformed

   private void copyMessageArea() {
      if (messageArea.getSelectionEnd() <= messageArea.getSelectionStart()) {
         messageArea.selectAll();
         messageArea.copy();
         messageArea.select(0, 0);
      } else {
         messageArea.copy();
      }
   }

   private void btnClearMessagesActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnClearMessagesActionPerformed
      messageArea.setText("");
   }//GEN-LAST:event_btnClearMessagesActionPerformed

   private void btnCutMessagesActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnCutMessagesActionPerformed
      cutMessageArea();
   }//GEN-LAST:event_btnCutMessagesActionPerformed

   private void cutMessageArea() {
      if (messageArea.getSelectionEnd() <= messageArea.getSelectionStart()) {
         messageArea.selectAll();
      }
      messageArea.cut();
   }

   private void messageAreaKeyTyped(java.awt.event.KeyEvent evt) {//GEN-FIRST:event_messageAreaKeyTyped
      evt.consume();
   }//GEN-LAST:event_messageAreaKeyTyped

   private void messageAreaKeyPressed(java.awt.event.KeyEvent evt) {//GEN-FIRST:event_messageAreaKeyPressed
      switch (evt.getKeyCode()) {
         case KeyEvent.VK_X:
            if (evt.isControlDown()) {
               cutMessageArea();
            }
            break;
         case KeyEvent.VK_C:
            if (evt.isControlDown()) {
               copyMessageArea();
            }
            break;
         case KeyEvent.VK_UP:
         case KeyEvent.VK_DOWN:
         case KeyEvent.VK_LEFT:
         case KeyEvent.VK_RIGHT:
         case KeyEvent.VK_PAGE_UP:
         case KeyEvent.VK_PAGE_DOWN:
         case KeyEvent.VK_END:
         case KeyEvent.VK_HOME:
            break;
         default:
            evt.consume();
            break;
      }
   }//GEN-LAST:event_messageAreaKeyPressed

   private void btnOpenActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnOpenActionPerformed
      final JFileChooser fileChooser;
      if (serversFile != null && serversFile.getParentFile() != null && serversFile.getParentFile().isDirectory()) {
         fileChooser = new JFileChooser(serversFile.getParentFile());
      } else {
         fileChooser = new JFileChooser(new File(System.getProperty("user.home")));
      }
      fileChooser.setFileFilter(new FileNameExtensionFilter("Arquivos ini", "ini"));
      if (fileChooser.showOpenDialog(this) == JFileChooser.APPROVE_OPTION) {
         serversFile = fileChooser.getSelectedFile();
         loadConfigurations();
      }
   }//GEN-LAST:event_btnOpenActionPerformed

   private void btnSaveAsActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnSaveAsActionPerformed
      final JFileChooser fileChooser;
      if (serversFile != null && serversFile.getParentFile() != null && serversFile.getParentFile().isDirectory()) {
         fileChooser = new JFileChooser(serversFile.getParentFile());
      } else {
         fileChooser = new JFileChooser(new File(System.getProperty("user.home")));
      }
      if (fileChooser.showSaveDialog(this) == JFileChooser.APPROVE_OPTION) {
         File saveFile = fileChooser.getSelectedFile();
         if (TerminalUtil.getFileExtension(saveFile).isBlank()) {
            saveFile = new File(saveFile.getAbsolutePath() + ".ini");
         }
         if (!saveFile.equals(serversFile)) {
            if (!saveFile.exists() || confirmsReplace(saveFile)) {
               serversFile = saveFile;
               saveConfiguration();
            }
         } else {
            saveConfiguration();
         }
      }
   }//GEN-LAST:event_btnSaveAsActionPerformed

   private void stopService(int rowModel) {
      final ServerTableModel model = serversModel();
      final RemoteTerminalServer server = model.getServerAt(rowModel);
      if (server == null || !server.isConnected()) {
         return;
      }
      if (server.isReadOnly()) {
         executor.execute(() -> message(TerminalUtil.getMessage("AdminClientWindow.readonly_not_allowed.msg", server)));
         return;
      }
      executor.execute(() -> message(TerminalUtil.getMessage("AdminClientWindow.stopping.msg", server)));
      executor.execute(() -> {
         try {
            server.stopTerminalEmulationService();
            message(TerminalUtil.getMessage("AdminClientWindow.done.msg"), "\n");
         } catch (Throwable ex) {
            message(TerminalUtil.getMessage("AdminClientWindow.error.msg"), "\n", ex.getMessage(),
                    "\n================================================================================\n");
            LOG.error("Error stopping server " + server.getConfiguration().address() + ":"
                    + server.getConfiguration().adminPort(), ex);
         } finally {
            model.notifyServerUpdate(rowModel);
         }
      });
   }

   private void startService(int rowModel) {
      final ServerTableModel model = serversModel();
      final RemoteTerminalServer server = model.getServerAt(rowModel);
      if (server == null || !server.isConnected()) {
         return;
      }
      if (server.isReadOnly()) {
         executor.execute(() -> message(TerminalUtil.getMessage("AdminClientWindow.readonly_not_allowed.msg", server)));
         return;
      }
      executor.execute(() -> message(TerminalUtil.getMessage("AdminClientWindow.starting.msg", server)));
      executor.execute(() -> {
         try {
            server.startTerminalEmulationService();
            message(TerminalUtil.getMessage("AdminClientWindow.done.msg"), "\n");
         } catch (Throwable ex) {
            message(TerminalUtil.getMessage("AdminClientWindow.error.msg"), "\n", ex.getMessage(),
                    "\n================================================================================\n");
            LOG.error("Error starting server " + server.getConfiguration().address() + ":"
                    + server.getConfiguration().adminPort(), ex);
         } finally {
            model.notifyServerUpdate(rowModel);
         }
      });
   }

   private void restartService(int rowModel) {
      final ServerTableModel model = serversModel();
      final RemoteTerminalServer server = model.getServerAt(rowModel);
      if (server == null || !server.isConnected()) {
         return;
      }
      if (server.isReadOnly()) {
         executor.execute(() -> message(TerminalUtil.getMessage("AdminClientWindow.readonly_not_allowed.msg", server)));
         return;
      }
      executor.execute(() -> message(TerminalUtil.getMessage("AdminClientWindow.restarting.msg", server)));
      executor.execute(() -> {
         try {
            server.stopTerminalEmulationService();
            server.startTerminalEmulationService();
            message(TerminalUtil.getMessage("AdminClientWindow.done.msg"), "\n");
         } catch (Throwable ex) {
            message(TerminalUtil.getMessage("AdminClientWindow.error.msg"), "\n", ex.getMessage(),
                    "\n================================================================================\n");
            LOG.error("Error restarting server " + server.getConfiguration().address() + ":"
                    + server.getConfiguration().adminPort(), ex);
         } finally {
            model.notifyServerUpdate(rowModel);
         }
      });
   }

   private void btnStopServiceActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnStopServiceActionPerformed
      final int[] selectedRows = modelSelectedRows();
      if (selectedRows.length > 0
              && JOptionPane.showConfirmDialog(this, TerminalUtil.getMessage("MainWindow.stopService.confirm.msg"),
                      TerminalUtil.getMessage("MainWindow.stopService.confirm.title"),
                      JOptionPane.YES_NO_OPTION, JOptionPane.QUESTION_MESSAGE) == 0) {
         for (int row : selectedRows) {
            stopService(row);
         }
         executor.execute(() -> verifySelection(modelSelectedRows()));
      }
   }//GEN-LAST:event_btnStopServiceActionPerformed

   private void btnStartServiceActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnStartServiceActionPerformed
      final int[] selectedRows = modelSelectedRows();
      for (int row : selectedRows) {
         startService(row);
      }
      executor.execute(() -> verifySelection(modelSelectedRows()));
   }//GEN-LAST:event_btnStartServiceActionPerformed

   private void btnRestartServiceActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnRestartServiceActionPerformed
      final int[] selectedRows = modelSelectedRows();
      for (int row : selectedRows) {
         restartService(row);
      }
      executor.execute(() -> verifySelection(modelSelectedRows()));
   }//GEN-LAST:event_btnRestartServiceActionPerformed

   private void btnKillSessionsActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnKillSessionsActionPerformed
      final int[] selectedRows = modelSelectedRows();
      if (selectedRows.length > 0
              && JOptionPane.showConfirmDialog(this, TerminalUtil.getMessage("AdminClientWindow.killingSessions.confirm.msg"),
                      TerminalUtil.getMessage("AdminClientWindow.killingSessions.confirm.title"),
                      JOptionPane.YES_NO_OPTION, JOptionPane.QUESTION_MESSAGE) == 0) {
         for (int row : selectedRows) {
            killSessions(row);
         }
         executor.execute(() -> verifySelection(modelSelectedRows()));
      }
   }//GEN-LAST:event_btnKillSessionsActionPerformed

   private void btnRemoteFilesActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnRemoteFilesActionPerformed
      final RemoteFilesWindow window;
      final ServerTableModel model = serversModel();
      final RemoteTerminalServer server = model.getServerAt(modelSelectedRow());
      if (server == null || !server.isConnected()) {
         return;
      }
      window = new RemoteFilesWindow(this, server);
      window.addListener(this);
      window.setVisible(true);
   }//GEN-LAST:event_btnRemoteFilesActionPerformed

   private void btnConfigServersActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnConfigServersActionPerformed
      final int[] selectedRows = modelSelectedRows();
      final ServerTableModel model = serversModel();
      final List<RemoteTerminalServer> selectedServers = new ArrayList<>(selectedRows.length);
      for (int row : selectedRows) {
         final RemoteTerminalServer server = model.getServerAt(row);
         if (server.isConnected()) {
            selectedServers.add(server);
         }
      }
      if (!selectedServers.isEmpty()) {
         final ServerConfigurationsWindow window = new ServerConfigurationsWindow(this, selectedServers, this);
         window.setVisible(true);
      }
   }//GEN-LAST:event_btnConfigServersActionPerformed

   private void btnSessionsActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnSessionsActionPerformed
      final int[] selectedRows = modelSelectedRows();
      for (int row : selectedRows) {
         showSessionsWindow(row);
      }
   }//GEN-LAST:event_btnSessionsActionPerformed

   private void showSessionsWindow(final int row) {
      final RemoteTerminalServer server = serversModel().getServerAt(row);
      if (server != null) {
         executor.execute(() -> {
            if (server.isConnected()) {
               SessionsWindow window = sessionsWindows.get(server);
               if (window != null) {
                  window.toFront();
               } else {
                  window = new SessionsWindow(this, server, executor);
                  sessionsWindows.put(server, window);
                  window.addWindowListener(this);
                  window.setVisible(true);
               }
            }
         });
      }
   }

   private void hideSessionsWindow(RemoteTerminalServer server) {
      final SessionsWindow window;
      if (server.isConnected()) {
         window = sessionsWindows.get(server);
      } else {
         window = sessionsWindows.remove(server);
      }
      if (window != null) {
         window.setVisible(false);
         window.dispose();
      }
   }

   private void killSessions(int rowModel) {
      final ServerTableModel model = serversModel();
      final RemoteTerminalServer server = model.getServerAt(rowModel);
      if (server == null || !server.isConnected()) {
         return;
      }
      if (server.isReadOnly()) {
         executor.execute(() -> message(TerminalUtil.getMessage("AdminClientWindow.readonly_not_allowed.msg", server)));
         return;
      }
      executor.execute(() -> message(TerminalUtil.getMessage("AdminClientWindow.killingSessions.msg", server)));
      executor.execute(() -> {
         try {
            final int sessionsKilled;
            batchOperation = true;
            sessionsKilled = server.killAllSessions();
            server.updateLocalState();
            server.updateSessions();
            message(TerminalUtil.getMessage("AdminClientWindow.done.msg"), ". ",
                    TerminalUtil.getMessage("AdminClientWindow.killedSessions.msg", sessionsKilled), ".\n");
         } catch (Throwable ex) {
            message(TerminalUtil.getMessage("AdminClientWindow.error.msg"), "\n", ex.getMessage(),
                    "\n================================================================================\n");
            LOG.error("Error killing session on server " + server.getConfiguration().address() + ":"
                    + server.getConfiguration().adminPort(), ex);
         } finally {
            updateConnectionsAndSessionsCount();
            batchOperation = false;
            model.notifyServerUpdate(rowModel);
         }
      });
   }

   private boolean confirmsReplace(final File file) {
      return JOptionPane.showConfirmDialog(this, TerminalUtil.getMessage("AdminClientWindow.replace_config.msg", file),
              TerminalUtil.getMessage("AdminClientWindow.replace_confirm.msg"),
              JOptionPane.YES_NO_OPTION, JOptionPane.QUESTION_MESSAGE) == 0;
   }

   private void showServerWindow(final int row) {
      final RemoteTerminalServer server = serversModel().getServerAt(row);
      if (server != null) {
         executor.execute(() -> {
            if (server.isConnected()) {
               ServerDetailsWindow window = monitorWindows.get(server);
               if (window != null) {
                  window.toFront();
               } else {
                  window = new ServerDetailsWindow(server);
                  monitorWindows.put(server, window);
                  window.addWindowListener(this);
                  window.setVisible(true);
               }
            }
         });
      }
   }

   private void hideServerWindow(RemoteTerminalServer server) {
      final ServerDetailsWindow window;
      if (server.isConnected()) {
         window = monitorWindows.get(server);
      } else {
         window = monitorWindows.remove(server);
      }
      if (window != null) {
         window.setVisible(false);
         window.dispose();
      }
   }

   /**
    * @param args the command line arguments
    */
   public static void main(String[] args) {
      /* Set the Nimbus look and feel */
      //<editor-fold defaultstate="collapsed" desc=" Look and feel setting code (optional) ">
      /* If Nimbus (introduced in Java SE 6) is not available, stay with the default look and feel.
         * For details see http://download.oracle.com/javase/tutorial/uiswing/lookandfeel/plaf.html
       */
      try {
         UIManager.setLookAndFeel(UIManager.getSystemLookAndFeelClassName());
      } catch (ClassNotFoundException | InstantiationException | IllegalAccessException | javax.swing.UnsupportedLookAndFeelException ex) {
         java.util.logging.Logger.getLogger(MainWindow.class.getName()).log(java.util.logging.Level.SEVERE, null, ex);
      }
      //</editor-fold>
      //</editor-fold>
      //</editor-fold>
      //</editor-fold>

      //</editor-fold>
      //</editor-fold>

      /* Create and display the form */
      java.awt.EventQueue.invokeLater(() -> new MainWindow().setVisible(true));
   }

   // Variables declaration - do not modify//GEN-BEGIN:variables
   private javax.swing.JTextField activeConnections;
   private javax.swing.JLabel activeConnectionsLabel;
   private javax.swing.JButton btnAdd;
   private javax.swing.JButton btnClearMessages;
   private javax.swing.JButton btnClose;
   private javax.swing.JButton btnConfigServers;
   private javax.swing.JButton btnConnect;
   private javax.swing.JButton btnCopyMessages;
   private javax.swing.JButton btnCutMessages;
   private javax.swing.JButton btnDisconnect;
   private javax.swing.JButton btnEdit;
   private javax.swing.JButton btnKillAdminSessions;
   private javax.swing.JButton btnKillSessions;
   private javax.swing.JButton btnOpen;
   private javax.swing.JButton btnRefresh;
   private javax.swing.JButton btnRemoteFiles;
   private javax.swing.JButton btnRemove;
   private javax.swing.JButton btnRestartService;
   private javax.swing.JButton btnSave;
   private javax.swing.JButton btnSaveAs;
   private javax.swing.JButton btnServer;
   private javax.swing.JButton btnSessions;
   private javax.swing.JButton btnStartService;
   private javax.swing.JButton btnStopService;
   private javax.swing.JPanel jPanel1;
   private javax.swing.JPanel jPanel2;
   private javax.swing.JPanel jPanel3;
   private javax.swing.JScrollPane jScrollPane1;
   private javax.swing.JScrollPane jScrollPane2;
   private javax.swing.JToolBar.Separator jSeparator1;
   private javax.swing.JToolBar.Separator jSeparator10;
   private javax.swing.JToolBar.Separator jSeparator2;
   private javax.swing.JToolBar.Separator jSeparator3;
   private javax.swing.JToolBar.Separator jSeparator4;
   private javax.swing.JToolBar.Separator jSeparator5;
   private javax.swing.JToolBar.Separator jSeparator6;
   private javax.swing.JToolBar.Separator jSeparator9;
   private javax.swing.JToolBar jToolBar1;
   private javax.swing.JToolBar jToolBar2;
   private javax.swing.JToolBar jToolBar3;
   private javax.swing.JTextArea messageArea;
   private javax.swing.JTable serversTable;
   private javax.swing.JLabel sessionsCountLabel;
   private javax.swing.JSplitPane splitPanel;
   private javax.swing.JTextField totalSessions;
   // End of variables declaration//GEN-END:variables

   private void loadConfigurations() {
      if (serversFile != null && serversFile.isFile()) {
         try {
            final ServerTableModel model = (ServerTableModel) serversTable.getModel();
            final Wini ini = new Wini(serversFile);
            closeAllConnections();
            model.clearServers();
            for (Section s : ini.values()) {
               if (s.containsKey("address")) {
                  final RemoteTerminalServer server;
                  server = new RemoteTerminalServer(s.getName(), s.get("address"), s.get("adminPort", Integer.class, 7656),
                          protocolProvider, this);
                  model.addServer(server);
                  server.addListener(this);
               }
            }
            serversTable.repaint();
            setTitle(TerminalUtil.getMessage("AdminClientWindow.servers_title.msg", serversFile.getName()) + localVersion());
         } catch (IOException | AuthenticatorException ex) {
            setTitle(TerminalUtil.getMessage("AdminClientWindow.title") + localVersion());
            JOptionPane.showMessageDialog(this, ex.getLocalizedMessage(),
                    TerminalUtil.getMessage("AdminClientWindow.error_loading.msg"), JOptionPane.ERROR_MESSAGE);
            LOG.error("Error reading configuration file.", ex);
         }
      }
   }

   private void closeAllConnections() {
      final ServerTableModel model = serversModel();
      for (int i = 0; i < model.getRowCount(); i++) {
         disconnectFromServer(i);
      }
      executor.execute(() -> verifySelection(new int[]{}));
   }

   @Override
   public void stateUpdate(TerminalServer server) {
      updateConnectionsAndSessionsCount();
   }

   @Override
   public void statusChanged(TerminalServer server, final ServerStatus previousStatus) {
      final RemoteTerminalServer remoteServer = (RemoteTerminalServer) server;
      final ServerTableModel model = serversModel();
      model.notifyServerUpdate(model.indexOfServer(remoteServer.getId()));
      if (ServerStatus.DISCONNECTED == remoteServer.getStatus()
              || ServerStatus.RUNNING == remoteServer.getStatus()
              || ServerStatus.STOPPED == remoteServer.getStatus()) {
         updateConnectionsAndSessionsCount();
      }
      verifySelection(modelSelectedRows());
   }

   private void updateConnectionsAndSessionsCount() {
      ServerTableModel model = (ServerTableModel) serversTable.getModel();
      int connectionsCount = 0;
      int sessionsCount = 0;
      for (int i = 0; i < serversTable.getRowCount(); i++) {
         final RemoteTerminalServer server = model.getServerAt(serversTable.convertRowIndexToModel(i));
         if (server != null && server.isConnected()) {
            ++connectionsCount;
            sessionsCount += server.getSessionsCount();
         }
      }
      activeConnections.setText(Integer.toString(connectionsCount));
      totalSessions.setText(Integer.toString(sessionsCount));
   }

   @Override
   public void sessionOpen(final Session session) {
      if (!batchOperation) {
         updateConnectionsAndSessionsCount();
      }
   }

   @Override
   public void sessionClose(final Session session) {
      if (!batchOperation) {
         updateConnectionsAndSessionsCount();
      }
   }

   @Override
   public void configurationLoaded(TerminalServer server) {
      if (server instanceof RemoteTerminalServer) {
         updateServerTableRow((RemoteTerminalServer) server);
      }
   }

   @Override
   public void serviceStart(TerminalServer server) {
      // ignore
   }

   @Override
   public void serviceStop(TerminalServer server) {
      // ignore
   }

   private void updateServerTableRow(RemoteTerminalServer server) {
      final ServerTableModel model = serversModel();
      final int row = model.indexOfServer(server.getId());
      if (row >= 0) {
         model.fireTableChanged(new TableModelEvent(model, row));
      }
   }

   @Override
   public void configurationSaved(TerminalServer server) {
      if (server instanceof RemoteTerminalServer) {
         updateServerTableRow((RemoteTerminalServer) server);
      }
   }

   @Override
   public void notification(TerminalServer server, String message) {
      message(message);
   }

   private void verifySelection(int[] selectedRows) {
      final ServerTableModel model = serversModel();
      int connectedCount = 0;
      int readOnlyCount = 0;
      int runningCount = 0;
      int stoppedCount = 0;
      for (int row : selectedRows) {
         final RemoteTerminalServer server = model.getServerAt(row);
         if (server.isConnected()) {
            ++connectedCount;
         }
         if (server.isReadOnly()) {
            ++readOnlyCount;
         }
         switch (server.getStatus()) {
            case RUNNING:
               ++runningCount;
               break;
            case STOPPED:
               ++stoppedCount;
               break;
         }
      }
      btnRemove.setEnabled(selectedRows.length > 0);
      btnEdit.setEnabled(selectedRows.length == 1 && connectedCount == 0);
      btnConnect.setEnabled(connectedCount < selectedRows.length && selectedRows.length > 0);
      btnKillAdminSessions.setEnabled(connectedCount < selectedRows.length && selectedRows.length > 0);
      btnServer.setEnabled(connectedCount > 0);
      btnDisconnect.setEnabled(connectedCount > 0);
      btnRefresh.setEnabled(connectedCount > 0);
      btnStopService.setEnabled(runningCount > 0 && readOnlyCount < connectedCount);
      btnStartService.setEnabled(stoppedCount > 0 && readOnlyCount < connectedCount);
      btnRestartService.setEnabled(runningCount > 0 && readOnlyCount < connectedCount);
      btnConfigServers.setEnabled(readOnlyCount < connectedCount);
      btnKillSessions.setEnabled(runningCount > 0 && readOnlyCount < connectedCount);
      btnRemoteFiles.setEnabled(selectedRows.length == 1 && connectedCount > 0);
   }

   @Override
   public void windowOpened(WindowEvent e) {
      // ignore
   }

   @Override
   public void windowClosing(WindowEvent e) {
      // ignore
   }

   @Override
   public void windowClosed(WindowEvent e) {
      TerminalServer server = ((ServerDetailsWindow) e.getWindow()).getServer();
      monitorWindows.remove(server);
      sessionsWindows.remove(server);
   }

   @Override
   public void windowIconified(WindowEvent e) {
      // ignore
   }

   @Override
   public void windowDeiconified(WindowEvent e) {
      // ignore
   }

   @Override
   public void windowActivated(WindowEvent e) {
      // ignore
   }

   @Override
   public void windowDeactivated(WindowEvent e) {
      // ignore
   }

   private void loadPreferences() {
      final Dimension tela = Toolkit.getDefaultToolkit().getScreenSize();
      final Properties prefs = PropertiesUtil.loadPropertiesFromFile(PREFERENCES_FILE);
      int width = PropertiesUtil.getProperty(prefs, "serversWindow.width", 486);
      int height = PropertiesUtil.getProperty(prefs, "serversWindow.height", 302);
      int row = PropertiesUtil.getProperty(prefs, "serversWindow.row", (tela.height - height) / 2);
      int col = PropertiesUtil.getProperty(prefs, "serversWindow.col", (tela.width - width) / 2);
      splitPanel.setDividerLocation(PropertiesUtil.getProperty(prefs, "serversWindow.divider", 250));
      setSize(width, height);
      setLocation(col, row);
      serversFile = new File(prefs.getProperty("serversList.name", DEFAULT_SERVERS_LIST_NAME));
   }

   private void savePreferences() {
      final Properties prefs = new Properties();
      prefs.setProperty("serversWindow.width", Integer.toString(getWidth()));
      prefs.setProperty("serversWindow.height", Integer.toString(getHeight()));
      prefs.setProperty("serversWindow.row", Integer.toString(getLocation().y));
      prefs.setProperty("serversWindow.col", Integer.toString(getLocation().x));
      prefs.setProperty("serversWindow.divider", Integer.toString(splitPanel.getDividerLocation()));
      prefs.setProperty("serversList.name", serversFile != null ? serversFile.getAbsolutePath() : DEFAULT_SERVERS_LIST_NAME);
      PropertiesUtil.savePropertiesToFile(prefs, PREFERENCES_FILE, false);
   }

   @Override
   public Credential getCredential(String id, boolean newCredential) {
      Credential credential = credentials.getOrDefault(id, lastCredential);
      if (credential == null || newCredential) {
         if (lastTriedCredential == null) {
            lastTriedCredential = new Credential(System.getProperty("user.name"), "");
         }
         credential = LoginWindow.showWindow(this, lastTriedCredential);
         if (credential != null) {
            credential.setPassword(Security.encrypt(credential.getPassword()));
            lastTriedCredential = credential;
         }
      }
      return credential;
   }

   @Override
   public void registerCredential(String id, Credential credential) {
      lastCredential = credential;
      if (credential != null) {
         credentials.put(id, credential);
      } else {
         credentials.remove(id);
      }
   }

   private String localVersion() {
      final String localVersion = TerminalUtil.getLocalVersion();
      if (!TerminalUtil.isEmpty(localVersion)) {
         return " - " + localVersion;
      }
      return "";
   }

   @Override
   public void valueChanged(ListSelectionEvent e) {
      if (!e.getValueIsAdjusting()) {
         verifySelection(modelSelectedRows());
      }
   }

   @Override
   public void listFiles(TerminalServer server, String folderPathname, List<FileInfo> files) {
      executor.execute(() -> message(TerminalUtil.getMessage("AdminClientWindow.remoteFile.list", files.size(), folderPathname,
              server), "\n"));
   }

   @Override
   public void removeFile(TerminalServer server, FileInfo file) {
      executor.execute(() -> message(TerminalUtil.getMessage("AdminClientWindow.remoteFile.remove", file, server), "\n"));
   }

   @Override
   public void uploadFile(TerminalServer server, FileInfo source, FileInfo target) {
      executor.execute(() -> message(TerminalUtil.getMessage("AdminClientWindow.remoteFile.upload", source, target, server), "\n"));
   }

   @Override
   public void downloadFile(TerminalServer server, FileInfo source, FileInfo target) {
      executor.execute(() -> message(TerminalUtil.getMessage("AdminClientWindow.remoteFile.download", source, server, target),
              "\n"));
   }

   @Override
   public void notification(final String msg) {
      executor.execute(() -> message(msg, "\n"));
   }
}
