/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.gui;

import org.rgt.RGTLogLevel;
import org.rgt.ServerStatus;
import org.rgt.TerminalException;
import org.rgt.TerminalServerConfiguration;
import org.rgt.auth.AuthenticatorException;
import java.awt.Dimension;
import java.awt.HeadlessException;
import java.awt.Toolkit;
import javax.swing.JOptionPane;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.rgt.TerminalServer;
import org.rgt.PropertiesUtil;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import java.io.File;
import java.net.InetAddress;
import java.net.NetworkInterface;
import java.net.SocketException;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Collections;
import java.util.Enumeration;
import java.util.Properties;
import javax.swing.DefaultComboBoxModel;
import javax.swing.JSpinner;
import javax.swing.event.ChangeEvent;
import javax.swing.event.ChangeListener;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.DefaultFormatter;
import org.rgt.CancelableExecutor;
import org.rgt.SerialExecutor;
import org.rgt.Session;
import org.rgt.TerminalServerListener;
import org.rgt.TerminalUtil;
import org.rgt.client.RemoteTerminalServer;
import org.rgt.options.TerminalOptionListener;

/**
 *
 * @author fabio_uggeri
 */
public class ServerDetailsWindow extends javax.swing.JFrame implements TerminalServerListener, TerminalOptionListener {

   private static final long serialVersionUID = 5293295767050071428L;

   private static final Logger LOG = LoggerFactory.getLogger(ServerDetailsWindow.class);

   private static final File PREFERENCES_FILE = new File(System.getProperty("user.home", "."), "rgt_preferences.properties");

   private RemoteTerminalServer server = null;

   private int configChangeCount = 0;

   private final CancelableExecutor executor;

   /**
    * Creates new form MainWindow
    *
    * @param server
    */
   public ServerDetailsWindow(final RemoteTerminalServer server) {
      this.executor = SerialExecutor.newInstance("Server: " + server.getId());
      this.server = server;
      try {
         initComponents();
         loadPreferences();
         configureServer();
      } catch (AuthenticatorException ex) {
         LOG.error("Error configuring server", ex);
         JOptionPane.showMessageDialog(null, ex.getMessage(),
                 TerminalUtil.getMessage("ServerMainWindow.error_config_server.msg"), JOptionPane.ERROR_MESSAGE);
      }
   }

   public TerminalServer getServer() {
      return server;
   }

   private void configureServer() throws AuthenticatorException {
      server.addListener(this);
      loadInterfaces();
      if (server.isEmbedded()) {
         loadConfiguration();
         loadUsers();
         if (!server.isRunningAdminService()) {
            try {
               server.startAdminService();
            } catch (TerminalException ex) {
               LOG.error("Error starting admin service", ex);
               JOptionPane.showMessageDialog(this, TerminalUtil.getMessage("ServerMainWindow.error_start_server.msg") + "\n"
                       + ex.getLocalizedMessage(), TerminalUtil.getMessage("ServerMainWindow.error_starting_admin.msg"),
                       JOptionPane.ERROR_MESSAGE);
            }
         }
      }
      configureFormFields();
      serverConfigurationToFormFields();
      configureComponents(server);
      server.getConfiguration().addListener(this);
      if (!TerminalUtil.isEmpty(server.getVersion())) {
         setTitle("Terminal Server - " + server.getVersion());
      } else {
         setTitle("Terminal Server");
      }
   }

   private void configureFormFields() {
      address.addItemListener(new EditionListener());
      port.addChangeListener(new EditionListener());
      ((DefaultFormatter) ((JSpinner.DefaultEditor) port.getEditor()).getTextField().getFormatter()).setCommitsOnValidEdit(true);
      serverLogPathName.getDocument().addDocumentListener(new EditionListener());
      serverLogLevel.addItemListener(new EditionListener());
      appLogPathName.getDocument().addDocumentListener(new EditionListener());
      appLogLevel.addItemListener(new EditionListener());
      teLogPathName.getDocument().addDocumentListener(new EditionListener());
      teLogLevel.addItemListener(new EditionListener());
      sessionsCount.setText(Integer.toString(server.getSessionsCount()));
   }

   private void loadConfiguration() {
      try {
         server.loadConfiguration();
      } catch (TerminalException ex) {
         LOG.error("Error loading server configuration.", ex);
         JOptionPane.showMessageDialog(null, ex.getLocalizedMessage(),
                 TerminalUtil.getMessage("ServerMainWindow.error_load_config.msg"), JOptionPane.ERROR_MESSAGE);
      }
   }

   private void loadUsers() throws HeadlessException {
      if (server.getUserRepository() != null) {
         try {
            server.getUserRepository().load();
         } catch (AuthenticatorException ex) {
            LOG.error("Error loading users frm file.", ex);
            JOptionPane.showMessageDialog(this, ex.getLocalizedMessage(),
                    TerminalUtil.getMessage("ServerMainWindow.error_load_users.msg"), JOptionPane.ERROR_MESSAGE);
         }
      }
   }

   /**
    * This method is called from within the constructor to initialize the form. WARNING: Do NOT modify this code. The content of
    * this method is always regenerated by the Form Editor.
    */
   @SuppressWarnings("unchecked")
   // <editor-fold defaultstate="collapsed" desc="Generated Code">//GEN-BEGIN:initComponents
   private void initComponents() {

      jPanel5 = new javax.swing.JPanel();
      listenPanel = new javax.swing.JPanel();
      addressLabel = new javax.swing.JLabel();
      portLabel = new javax.swing.JLabel();
      port = new javax.swing.JSpinner();
      address = new javax.swing.JComboBox<>();
      jPanel6 = new javax.swing.JPanel();
      jPanel3 = new javax.swing.JPanel();
      statusBarStatusLabel = new javax.swing.JLabel();
      statusBarSessionLabel = new javax.swing.JLabel();
      status = new javax.swing.JLabel();
      sessionsCount = new javax.swing.JTextField();
      logPanel = new javax.swing.JPanel();
      logTemrinalPanel = new javax.swing.JPanel();
      terminalLogLevelLabel = new javax.swing.JLabel();
      teLogLevel = new javax.swing.JComboBox<>();
      terminalLogFileLabel = new javax.swing.JLabel();
      teLogPathName = new javax.swing.JTextField();
      logAppPanel = new javax.swing.JPanel();
      appLogLevelLabel = new javax.swing.JLabel();
      appLogLevel = new javax.swing.JComboBox<>();
      appLogFileLabel = new javax.swing.JLabel();
      appLogPathName = new javax.swing.JTextField();
      logServerPanel = new javax.swing.JPanel();
      serverLogLevelLabel = new javax.swing.JLabel();
      serverLogLevel = new javax.swing.JComboBox<>();
      serverLogFileLabel = new javax.swing.JLabel();
      serverLogPathName = new javax.swing.JTextField();
      jPanel1 = new javax.swing.JPanel();
      jToolBar1 = new javax.swing.JToolBar();
      btnStart = new javax.swing.JButton();
      jSeparator1 = new javax.swing.JToolBar.Separator();
      btnConfirmConfig = new javax.swing.JButton();
      btnRefresh = new javax.swing.JButton();
      btnUsers = new javax.swing.JButton();
      btnSessions = new javax.swing.JButton();
      jSeparator3 = new javax.swing.JToolBar.Separator();
      btnConfig = new javax.swing.JButton();
      jSeparator2 = new javax.swing.JToolBar.Separator();
      btnSave = new javax.swing.JButton();
      jPanel7 = new javax.swing.JPanel();
      jToolBar2 = new javax.swing.JToolBar();
      btnQuit = new javax.swing.JButton();

      setDefaultCloseOperation(javax.swing.WindowConstants.DO_NOTHING_ON_CLOSE);
      setTitle("Terminal Server");
      setIconImage(new javax.swing.ImageIcon(getClass().getResource("/computer.png")).getImage());
      addWindowListener(new java.awt.event.WindowAdapter() {
         public void windowClosed(java.awt.event.WindowEvent evt) {
            formWindowClosed(evt);
         }
      });

      java.util.ResourceBundle bundle = java.util.ResourceBundle.getBundle("org/rgt/gui/Bundle"); // NOI18N
      listenPanel.setBorder(javax.swing.BorderFactory.createTitledBorder(bundle.getString("ServerMainWindow.listenPanel.label"))); // NOI18N
      listenPanel.setToolTipText("");

      addressLabel.setText(bundle.getString("ServerMainWindow.listenPanel.address.label")); // NOI18N

      portLabel.setText(bundle.getString("ServerMainWindow.listenPanel.port.label")); // NOI18N

      port.setModel(new javax.swing.SpinnerNumberModel(7654, 1, 65535, 1));

      address.setModel(new javax.swing.DefaultComboBoxModel<>(new String[] { "Default" }));

      javax.swing.GroupLayout listenPanelLayout = new javax.swing.GroupLayout(listenPanel);
      listenPanel.setLayout(listenPanelLayout);
      listenPanelLayout.setHorizontalGroup(
         listenPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(listenPanelLayout.createSequentialGroup()
            .addGroup(listenPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
               .addComponent(addressLabel, javax.swing.GroupLayout.Alignment.TRAILING)
               .addComponent(portLabel, javax.swing.GroupLayout.Alignment.TRAILING))
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.RELATED)
            .addGroup(listenPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
               .addComponent(port)
               .addComponent(address, 0, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)))
      );
      listenPanelLayout.setVerticalGroup(
         listenPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(listenPanelLayout.createSequentialGroup()
            .addGroup(listenPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.CENTER)
               .addComponent(address, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE)
               .addComponent(addressLabel))
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.RELATED)
            .addGroup(listenPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.CENTER)
               .addComponent(port, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE)
               .addComponent(portLabel))
            .addContainerGap())
      );

      javax.swing.GroupLayout jPanel5Layout = new javax.swing.GroupLayout(jPanel5);
      jPanel5.setLayout(jPanel5Layout);
      jPanel5Layout.setHorizontalGroup(
         jPanel5Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addComponent(listenPanel, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
      );
      jPanel5Layout.setVerticalGroup(
         jPanel5Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addComponent(listenPanel, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
      );

      jPanel3.setBorder(javax.swing.BorderFactory.createEtchedBorder());

      statusBarStatusLabel.setText(bundle.getString("ServerMainWindow.statusBar.status.label")); // NOI18N

      statusBarSessionLabel.setText(bundle.getString("ServerMainWindow.statusBar.session.label")); // NOI18N

      status.setFont(new java.awt.Font("Tahoma", 1, 11)); // NOI18N
      status.setText("Parado");

      sessionsCount.setEditable(false);
      sessionsCount.setHorizontalAlignment(javax.swing.JTextField.RIGHT);
      sessionsCount.setBorder(null);
      sessionsCount.setFocusable(false);

      javax.swing.GroupLayout jPanel3Layout = new javax.swing.GroupLayout(jPanel3);
      jPanel3.setLayout(jPanel3Layout);
      jPanel3Layout.setHorizontalGroup(
         jPanel3Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(jPanel3Layout.createSequentialGroup()
            .addContainerGap()
            .addComponent(statusBarStatusLabel)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.RELATED)
            .addComponent(status, javax.swing.GroupLayout.DEFAULT_SIZE, 285, Short.MAX_VALUE)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.UNRELATED)
            .addComponent(statusBarSessionLabel)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.RELATED)
            .addComponent(sessionsCount, javax.swing.GroupLayout.PREFERRED_SIZE, 87, javax.swing.GroupLayout.PREFERRED_SIZE)
            .addContainerGap())
      );
      jPanel3Layout.setVerticalGroup(
         jPanel3Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(javax.swing.GroupLayout.Alignment.TRAILING, jPanel3Layout.createSequentialGroup()
            .addGap(0, 1, Short.MAX_VALUE)
            .addGroup(jPanel3Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.BASELINE)
               .addComponent(statusBarStatusLabel)
               .addComponent(statusBarSessionLabel)
               .addComponent(status)
               .addComponent(sessionsCount, javax.swing.GroupLayout.PREFERRED_SIZE, 17, javax.swing.GroupLayout.PREFERRED_SIZE)))
      );

      logPanel.setBorder(javax.swing.BorderFactory.createTitledBorder(bundle.getString("ServerMainWindow.logPanel.label"))); // NOI18N

      logTemrinalPanel.setBorder(javax.swing.BorderFactory.createTitledBorder(bundle.getString("ServerMainWindow.logPanel.terminalPanel.label"))); // NOI18N

      terminalLogLevelLabel.setText(bundle.getString("ServerMainWindow.logPanel.terminalPanel.level.label")); // NOI18N

      teLogLevel.setModel(new javax.swing.DefaultComboBoxModel<>(new String[] { "TRACE", "DEBUG", "INFO", "WARNING", "ERROR", "OFF" }));

      terminalLogFileLabel.setText(bundle.getString("ServerMainWindow.logPanel.terminalPanel.file.label")); // NOI18N

      javax.swing.GroupLayout logTemrinalPanelLayout = new javax.swing.GroupLayout(logTemrinalPanel);
      logTemrinalPanel.setLayout(logTemrinalPanelLayout);
      logTemrinalPanelLayout.setHorizontalGroup(
         logTemrinalPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(logTemrinalPanelLayout.createSequentialGroup()
            .addContainerGap()
            .addComponent(terminalLogLevelLabel)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.UNRELATED)
            .addComponent(teLogLevel, 0, 99, Short.MAX_VALUE)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.UNRELATED)
            .addComponent(terminalLogFileLabel)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.UNRELATED)
            .addComponent(teLogPathName, javax.swing.GroupLayout.DEFAULT_SIZE, 276, Short.MAX_VALUE)
            .addContainerGap())
      );
      logTemrinalPanelLayout.setVerticalGroup(
         logTemrinalPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(logTemrinalPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.CENTER)
            .addComponent(terminalLogLevelLabel)
            .addComponent(teLogLevel, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE)
            .addComponent(terminalLogFileLabel)
            .addComponent(teLogPathName, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE))
      );

      logAppPanel.setBorder(javax.swing.BorderFactory.createTitledBorder(bundle.getString("ServerMainWindow.logPanel.appPanel.label"))); // NOI18N

      appLogLevelLabel.setText(bundle.getString("ServerMainWindow.logPanel.appPanel.level.label")); // NOI18N

      appLogLevel.setModel(new javax.swing.DefaultComboBoxModel<>(new String[] { "TRACE", "DEBUG", "INFO", "WARNING", "ERROR", "OFF" }));

      appLogFileLabel.setText(bundle.getString("ServerMainWindow.logPanel.appPanel.file.label")); // NOI18N

      javax.swing.GroupLayout logAppPanelLayout = new javax.swing.GroupLayout(logAppPanel);
      logAppPanel.setLayout(logAppPanelLayout);
      logAppPanelLayout.setHorizontalGroup(
         logAppPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(logAppPanelLayout.createSequentialGroup()
            .addContainerGap()
            .addComponent(appLogLevelLabel)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.UNRELATED)
            .addComponent(appLogLevel, 0, 99, Short.MAX_VALUE)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.UNRELATED)
            .addComponent(appLogFileLabel)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.UNRELATED)
            .addComponent(appLogPathName, javax.swing.GroupLayout.DEFAULT_SIZE, 276, Short.MAX_VALUE)
            .addContainerGap())
      );
      logAppPanelLayout.setVerticalGroup(
         logAppPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(logAppPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.BASELINE)
            .addComponent(appLogLevel, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE)
            .addComponent(appLogLevelLabel)
            .addComponent(appLogFileLabel)
            .addComponent(appLogPathName, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE))
      );

      logServerPanel.setBorder(javax.swing.BorderFactory.createTitledBorder(bundle.getString("ServerMainWindow.logPanel.serverPanel.label"))); // NOI18N

      serverLogLevelLabel.setText(bundle.getString("ServerMainWindow.logPanel.serverPanel.level.label")); // NOI18N

      serverLogLevel.setModel(new javax.swing.DefaultComboBoxModel<>(new String[] { "TRACE", "DEBUG", "INFO", "WARNING", "ERROR", "OFF" }));

      serverLogFileLabel.setText(bundle.getString("ServerMainWindow.logPanel.serverPanel.file.label")); // NOI18N

      javax.swing.GroupLayout logServerPanelLayout = new javax.swing.GroupLayout(logServerPanel);
      logServerPanel.setLayout(logServerPanelLayout);
      logServerPanelLayout.setHorizontalGroup(
         logServerPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(logServerPanelLayout.createSequentialGroup()
            .addContainerGap()
            .addComponent(serverLogLevelLabel)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.UNRELATED)
            .addComponent(serverLogLevel, 0, 99, Short.MAX_VALUE)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.UNRELATED)
            .addComponent(serverLogFileLabel)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.UNRELATED)
            .addComponent(serverLogPathName, javax.swing.GroupLayout.DEFAULT_SIZE, 276, Short.MAX_VALUE)
            .addContainerGap())
      );
      logServerPanelLayout.setVerticalGroup(
         logServerPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(logServerPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.BASELINE)
            .addComponent(serverLogLevel, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE)
            .addComponent(serverLogLevelLabel)
            .addComponent(serverLogFileLabel)
            .addComponent(serverLogPathName, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE))
      );

      javax.swing.GroupLayout logPanelLayout = new javax.swing.GroupLayout(logPanel);
      logPanel.setLayout(logPanelLayout);
      logPanelLayout.setHorizontalGroup(
         logPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addComponent(logTemrinalPanel, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
         .addComponent(logAppPanel, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
         .addComponent(logServerPanel, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
      );
      logPanelLayout.setVerticalGroup(
         logPanelLayout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(logPanelLayout.createSequentialGroup()
            .addComponent(logTemrinalPanel, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.RELATED)
            .addComponent(logAppPanel, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.RELATED)
            .addComponent(logServerPanel, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE))
      );

      javax.swing.GroupLayout jPanel6Layout = new javax.swing.GroupLayout(jPanel6);
      jPanel6.setLayout(jPanel6Layout);
      jPanel6Layout.setHorizontalGroup(
         jPanel6Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addComponent(jPanel3, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
         .addComponent(logPanel, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
      );
      jPanel6Layout.setVerticalGroup(
         jPanel6Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(jPanel6Layout.createSequentialGroup()
            .addComponent(logPanel, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE)
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.RELATED)
            .addComponent(jPanel3, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE))
      );

      jToolBar1.setBorder(javax.swing.BorderFactory.createEmptyBorder(1, 1, 1, 1));
      jToolBar1.setRollover(true);

      btnStart.setIcon(new javax.swing.ImageIcon(getClass().getResource("/start.png"))); // NOI18N
      btnStart.setToolTipText(bundle.getString("ServerMainWindow.btnStart.toolTipText")); // NOI18N
      btnStart.setDefaultCapable(false);
      btnStart.setFocusable(false);
      btnStart.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnStartActionPerformed(evt);
         }
      });
      jToolBar1.add(btnStart);
      jToolBar1.add(jSeparator1);

      btnConfirmConfig.setIcon(new javax.swing.ImageIcon(getClass().getResource("/ok.png"))); // NOI18N
      btnConfirmConfig.setToolTipText(bundle.getString("ServerMainWindow.btnConfirmConfig.toolTipText")); // NOI18N
      btnConfirmConfig.setDefaultCapable(false);
      btnConfirmConfig.setEnabled(false);
      btnConfirmConfig.setFocusable(false);
      btnConfirmConfig.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnConfirmConfig.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnConfirmConfig.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnConfirmConfigActionPerformed(evt);
         }
      });
      jToolBar1.add(btnConfirmConfig);

      btnRefresh.setIcon(new javax.swing.ImageIcon(getClass().getResource("/refresh2.png"))); // NOI18N
      btnRefresh.setToolTipText(bundle.getString("ServerMainWindow.btnRefresh.toolTipText")); // NOI18N
      btnRefresh.setDefaultCapable(false);
      btnRefresh.setEnabled(false);
      btnRefresh.setFocusable(false);
      btnRefresh.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnRefresh.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnRefresh.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnRefreshActionPerformed(evt);
         }
      });
      jToolBar1.add(btnRefresh);

      btnUsers.setIcon(new javax.swing.ImageIcon(getClass().getResource("/users.png"))); // NOI18N
      btnUsers.setToolTipText(bundle.getString("ServerMainWindow.btnUsers.toolTipText")); // NOI18N
      btnUsers.setDefaultCapable(false);
      btnUsers.setFocusable(false);
      btnUsers.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnUsersActionPerformed(evt);
         }
      });
      jToolBar1.add(btnUsers);

      btnSessions.setIcon(new javax.swing.ImageIcon(getClass().getResource("/sessions.png"))); // NOI18N
      btnSessions.setToolTipText(bundle.getString("ServerMainWindow.btnSessions.toolTipText")); // NOI18N
      btnSessions.setDefaultCapable(false);
      btnSessions.setEnabled(false);
      btnSessions.setFocusable(false);
      btnSessions.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnSessions.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnSessions.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnSessionsActionPerformed(evt);
         }
      });
      jToolBar1.add(btnSessions);
      jToolBar1.add(jSeparator3);

      btnConfig.setIcon(new javax.swing.ImageIcon(getClass().getResource("/config.png"))); // NOI18N
      btnConfig.setToolTipText(bundle.getString("ServerMainWindow.btnConfig.toolTipText")); // NOI18N
      btnConfig.setDefaultCapable(false);
      btnConfig.setFocusable(false);
      btnConfig.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnConfig.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnConfig.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnConfigActionPerformed(evt);
         }
      });
      jToolBar1.add(btnConfig);
      jToolBar1.add(jSeparator2);

      btnSave.setIcon(new javax.swing.ImageIcon(getClass().getResource("/diskette.png"))); // NOI18N
      btnSave.setToolTipText(bundle.getString("ServerMainWindow.btnSave.toolTipText")); // NOI18N
      btnSave.setDefaultCapable(false);
      btnSave.setFocusable(false);
      btnSave.setHorizontalTextPosition(javax.swing.SwingConstants.CENTER);
      btnSave.setVerticalTextPosition(javax.swing.SwingConstants.BOTTOM);
      btnSave.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnSaveActionPerformed(evt);
         }
      });
      jToolBar1.add(btnSave);

      javax.swing.GroupLayout jPanel1Layout = new javax.swing.GroupLayout(jPanel1);
      jPanel1.setLayout(jPanel1Layout);
      jPanel1Layout.setHorizontalGroup(
         jPanel1Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addComponent(jToolBar1, javax.swing.GroupLayout.DEFAULT_SIZE, 416, Short.MAX_VALUE)
      );
      jPanel1Layout.setVerticalGroup(
         jPanel1Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addComponent(jToolBar1, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
      );

      jToolBar2.setBorder(javax.swing.BorderFactory.createEmptyBorder(1, 1, 1, 1));
      jToolBar2.setRollover(true);

      btnQuit.setIcon(new javax.swing.ImageIcon(getClass().getResource("/exit.png"))); // NOI18N
      btnQuit.setToolTipText(bundle.getString("ServerMainWindow.btnQuit.toolTipText")); // NOI18N
      btnQuit.setDefaultCapable(false);
      btnQuit.setFocusable(false);
      btnQuit.addActionListener(new java.awt.event.ActionListener() {
         public void actionPerformed(java.awt.event.ActionEvent evt) {
            btnQuitActionPerformed(evt);
         }
      });
      jToolBar2.add(btnQuit);

      javax.swing.GroupLayout jPanel7Layout = new javax.swing.GroupLayout(jPanel7);
      jPanel7.setLayout(jPanel7Layout);
      jPanel7Layout.setHorizontalGroup(
         jPanel7Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(javax.swing.GroupLayout.Alignment.TRAILING, jPanel7Layout.createSequentialGroup()
            .addGap(0, 0, Short.MAX_VALUE)
            .addComponent(jToolBar2, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE))
      );
      jPanel7Layout.setVerticalGroup(
         jPanel7Layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addComponent(jToolBar2, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
      );

      javax.swing.GroupLayout layout = new javax.swing.GroupLayout(getContentPane());
      getContentPane().setLayout(layout);
      layout.setHorizontalGroup(
         layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addComponent(jPanel5, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
         .addGroup(layout.createSequentialGroup()
            .addComponent(jPanel1, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
            .addGap(28, 28, 28)
            .addComponent(jPanel7, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE))
         .addComponent(jPanel6, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
      );
      layout.setVerticalGroup(
         layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING)
         .addGroup(javax.swing.GroupLayout.Alignment.TRAILING, layout.createSequentialGroup()
            .addGroup(layout.createParallelGroup(javax.swing.GroupLayout.Alignment.LEADING, false)
               .addComponent(jPanel1, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE)
               .addComponent(jPanel7, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, Short.MAX_VALUE))
            .addPreferredGap(javax.swing.LayoutStyle.ComponentPlacement.RELATED)
            .addComponent(jPanel5, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE)
            .addGap(0, 0, Short.MAX_VALUE)
            .addComponent(jPanel6, javax.swing.GroupLayout.PREFERRED_SIZE, javax.swing.GroupLayout.DEFAULT_SIZE, javax.swing.GroupLayout.PREFERRED_SIZE))
      );

      pack();
   }// </editor-fold>//GEN-END:initComponents

   private void btnStartActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnStartActionPerformed
      try {
         final ServerStatus currentStatus = server.getStatus();
         server.updateLocalStatus();
         if (currentStatus == server.getStatus()) {
            switch (currentStatus) {
               case STOPPED:
                  startService();
                  break;
               case RUNNING:
                  stopTerminalEmulationService();
                  break;
               default:
                  break;
            }
         } else {
            refreshServerState();
            LOG.debug("Local server state refreshed.");
            JOptionPane.showMessageDialog(this, TerminalUtil.getMessage("ServerMainWindow.remote_updated.msg"),
                    "Remote state changed.", JOptionPane.INFORMATION_MESSAGE);
         }
      } catch (TerminalException ex) {
         LOG.error("Error getting server status.", ex);
         JOptionPane.showMessageDialog(this, ex.getLocalizedMessage(),
                 TerminalUtil.getMessage("ServerMainWindow.error_getting_status.msg"), JOptionPane.ERROR_MESSAGE);
         if (!server.isEmbedded() && !server.isConnected()) {
            exit();
         }
      }
   }//GEN-LAST:event_btnStartActionPerformed

   private void stopTerminalEmulationService() {
      if (connectionLost()) {
         return;
      }
      try {
         server.stopTerminalEmulationService();
      } catch (TerminalException ex) {
         LOG.error("Error stopping service", ex);
         JOptionPane.showMessageDialog(this, ex.getLocalizedMessage(),
                 TerminalUtil.getMessage("ServerMainWindow.error_stop_server.msg"), JOptionPane.ERROR_MESSAGE);
         if (!server.isEmbedded() && !server.isConnected()) {
            exit();
         }
      }
   }

   private boolean connectionLost() throws HeadlessException {
      if (!server.isEmbedded() && !server.isConnected()) {
         LOG.error("Connection lost with remote server.");
         JOptionPane.showMessageDialog(this, TerminalUtil.getMessage("ServerMainWindow.conn_lost_server.msg"),
                 TerminalUtil.getMessage("ServerMainWindow.conn_lost.msg"), JOptionPane.ERROR_MESSAGE);
         exit();
         return true;
      }
      return false;
   }

   private void btnUsersActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnUsersActionPerformed
      if (connectionLost()) {
         return;
      }
      new UsersWindow(this, server).setVisible(true);
   }//GEN-LAST:event_btnUsersActionPerformed

   private void exit() {
      savePreferences();
      executor.stop(true);
      setVisible(false);
      dispose();
      if (server.isEmbedded()) {
         if (server.isRunning()) {
            stopTerminalEmulationService();
         }
         if (server.isRunningAdminService()) {
            try {
               server.stopAdminService();
            } catch (TerminalException ex) {
               LOG.error("Error stopping admin service", ex);
            }
         }
         System.exit(0);
      }
   }

   private void btnQuitActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnQuitActionPerformed
      exit();
   }//GEN-LAST:event_btnQuitActionPerformed

   private void formWindowClosed(java.awt.event.WindowEvent evt) {//GEN-FIRST:event_formWindowClosed
      exit();
   }//GEN-LAST:event_formWindowClosed

   private void btnSaveActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnSaveActionPerformed
      if (connectionLost()) {
         return;
      }
      if (JOptionPane.showConfirmDialog(null, TerminalUtil.getMessage("ServerMainWindow.confirm_save.msg"),
              TerminalUtil.getMessage("ServerMainWindow.save_config.msg"), JOptionPane.YES_NO_OPTION,
              JOptionPane.QUESTION_MESSAGE) == 0) {
         try {
            formFieldsToServerConfiguration();
            server.saveConfiguration();
         } catch (TerminalException ex) {
            LOG.error("Error writing configuration file rgt_config.properties.", ex);
            JOptionPane.showMessageDialog(this, ex.getLocalizedMessage(),
                    TerminalUtil.getMessage("ServerMainWindow.error_saving.msg"), JOptionPane.ERROR_MESSAGE);
         }
      }
   }//GEN-LAST:event_btnSaveActionPerformed

   private void btnSessionsActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnSessionsActionPerformed
      if (!connectionLost()) {
         new SessionsWindow(this, server, executor).setVisible(true);
         executor.clear();
      }
   }//GEN-LAST:event_btnSessionsActionPerformed

   private void btnRefreshActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnRefreshActionPerformed
      executor.execute(this::refreshServerState);
   }//GEN-LAST:event_btnRefreshActionPerformed

   private void refreshServerState() throws HeadlessException {
      if (!server.isEmbedded() && !connectionLost()) {
         try {
            server.updateLocalState();
            sessionsCount.setText(Integer.toString(server.getSessionsCount()));
            serverConfigurationToFormFields();
         } catch (TerminalException ex) {
            LOG.error("Error updating server status.");
            JOptionPane.showMessageDialog(this, TerminalUtil.getMessage("ServerMainWindow.error_update_status.msg"),
                    TerminalUtil.getMessage("ServerMainWindow.error_updating_status.msg"), JOptionPane.ERROR_MESSAGE);
            if (!server.isConnected()) {
               exit();
            }
         }
      }
   }

   private void btnConfirmConfigActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnConfirmConfigActionPerformed
      try {
         formFieldsToServerConfiguration();
         server.updateRemoteConfiguration();
      } catch (TerminalException ex) {
         LOG.error("Error setting server configuration.");
         JOptionPane.showMessageDialog(this, ex.getMessage(), TerminalUtil.getMessage("ServerMainWindow.error_setting_config.msg"),
                 JOptionPane.ERROR_MESSAGE);
         if (!server.isEmbedded() && !server.isConnected()) {
            exit();
         }
      }
   }//GEN-LAST:event_btnConfirmConfigActionPerformed

   private void btnConfigActionPerformed(java.awt.event.ActionEvent evt) {//GEN-FIRST:event_btnConfigActionPerformed
      final ServerConfigurationsWindow window = new ServerConfigurationsWindow(this, Collections.singletonList(server), null);
      window.setVisible(true);
   }//GEN-LAST:event_btnConfigActionPerformed

   public void startService() {
      if (connectionLost()) {
         return;
      }
      if (!validConfiguration()) {
         return;
      }
      try {
         sessionsCount.setText("0");
         formFieldsToServerConfiguration();
         server.startTerminalEmulationService();
      } catch (TerminalException ex) {
         LOG.error("Error starting server.", ex);
         JOptionPane.showMessageDialog(this, ex.getMessage(), TerminalUtil.getMessage("ServerMainWindow.error_starting.msg"),
                 JOptionPane.ERROR_MESSAGE);
         if (!server.isEmbedded() && !server.isConnected()) {
            exit();
         }
      }
   }

   private void formFieldsToServerConfiguration() {
      final TerminalServerConfiguration config = server.getConfiguration();
      final String oldAddress = config.address().value();
      if (address.getSelectedIndex() == 0) {
         config.address().value("");
      } else {
         config.address().value(address.getSelectedItem().toString());
      }
      config.port().value((Integer) port.getValue());
      config.teLogPathName().value(teLogPathName.getText());
      config.teLogLevel().value(RGTLogLevel.valueOf(teLogLevel.getSelectedItem().toString()));
      config.serverLogPathName().value(serverLogPathName.getText());
      config.serverLogLevel().value(RGTLogLevel.valueOf(serverLogLevel.getSelectedItem().toString()));
      config.serverLogLevel().value().active();
      config.appLogPathName().value(appLogPathName.getText());
      config.appLogLevel().value(RGTLogLevel.valueOf(appLogLevel.getSelectedItem().toString()));
      configChangeCount = 0;
      btnConfirmConfig.setEnabled(false);
      if (server.isEmbedded() && server.isRunningAdminService() && !isValueEquals(oldAddress, config.address().value())) {
         try {
            server.stopAdminService();
            server.startAdminService();
         } catch (TerminalException ex) {
            LOG.error("Connection restarting admin service.");
            JOptionPane.showMessageDialog(this, ex.getMessage(),
                    TerminalUtil.getMessage("ServerMainWindow.error_starting_admin.msg"), JOptionPane.ERROR_MESSAGE);
         }
      }
   }

   private boolean validConfiguration() throws HeadlessException {
      final Collection<String> errors = server.getConfiguration().validate();
      if (!errors.isEmpty()) {
         final StringBuilder sb = new StringBuilder();
         for (String s : errors) {
            if (sb.length() > 0) {
               sb.append('\n');
            }
            sb.append(s);
         }
         LOG.error("Invalid configuration:\n{}", sb);
         JOptionPane.showMessageDialog(this, sb, TerminalUtil.getMessage("ServerMainWindow.invalid_config.msg"),
                 JOptionPane.WARNING_MESSAGE);
         return false;
      }
      return true;
   }

   // Variables declaration - do not modify//GEN-BEGIN:variables
   private javax.swing.JComboBox<String> address;
   private javax.swing.JLabel addressLabel;
   private javax.swing.JLabel appLogFileLabel;
   private javax.swing.JComboBox<String> appLogLevel;
   private javax.swing.JLabel appLogLevelLabel;
   private javax.swing.JTextField appLogPathName;
   private javax.swing.JButton btnConfig;
   private javax.swing.JButton btnConfirmConfig;
   private javax.swing.JButton btnQuit;
   private javax.swing.JButton btnRefresh;
   private javax.swing.JButton btnSave;
   private javax.swing.JButton btnSessions;
   private javax.swing.JButton btnStart;
   private javax.swing.JButton btnUsers;
   private javax.swing.JPanel jPanel1;
   private javax.swing.JPanel jPanel3;
   private javax.swing.JPanel jPanel5;
   private javax.swing.JPanel jPanel6;
   private javax.swing.JPanel jPanel7;
   private javax.swing.JToolBar.Separator jSeparator1;
   private javax.swing.JToolBar.Separator jSeparator2;
   private javax.swing.JToolBar.Separator jSeparator3;
   private javax.swing.JToolBar jToolBar1;
   private javax.swing.JToolBar jToolBar2;
   private javax.swing.JPanel listenPanel;
   private javax.swing.JPanel logAppPanel;
   private javax.swing.JPanel logPanel;
   private javax.swing.JPanel logServerPanel;
   private javax.swing.JPanel logTemrinalPanel;
   private javax.swing.JSpinner port;
   private javax.swing.JLabel portLabel;
   private javax.swing.JLabel serverLogFileLabel;
   private javax.swing.JComboBox<String> serverLogLevel;
   private javax.swing.JLabel serverLogLevelLabel;
   private javax.swing.JTextField serverLogPathName;
   private javax.swing.JTextField sessionsCount;
   private javax.swing.JLabel status;
   private javax.swing.JLabel statusBarSessionLabel;
   private javax.swing.JLabel statusBarStatusLabel;
   private javax.swing.JComboBox<String> teLogLevel;
   private javax.swing.JTextField teLogPathName;
   private javax.swing.JLabel terminalLogFileLabel;
   private javax.swing.JLabel terminalLogLevelLabel;
   // End of variables declaration//GEN-END:variables

   private void configureComponents(TerminalServer server) {
      final ServerStatus serverStatus = server.getStatus();
      final boolean stopped = ServerStatus.STOPPED == serverStatus;
      final boolean running = ServerStatus.RUNNING == serverStatus;
      btnConfirmConfig.setEnabled(configChangeCount > 0 && !server.isReadOnly());
      btnRefresh.setEnabled(!server.isEmbedded());
      btnSessions.setEnabled(running);
      btnSave.setEnabled(stopped && !server.isReadOnly());
      btnUsers.setEnabled(server.getUserRepository() != null);
      address.setEnabled(stopped && !server.isReadOnly());
      port.setEnabled(stopped && !server.isReadOnly());
      appLogPathName.setEnabled(stopped && !server.isReadOnly());
      teLogPathName.setEnabled(stopped && !server.isReadOnly());
      serverLogPathName.setEnabled(stopped && !server.isReadOnly());
      status.setText(server.getStatus().toString());
      switch (serverStatus) {
         case RUNNING:
            btnStart.setIcon(new javax.swing.ImageIcon(getClass().getResource("/stop.png")));
            btnStart.setToolTipText("Parar");
            break;
         case STOPPED:
            btnStart.setIcon(new javax.swing.ImageIcon(getClass().getResource("/start.png")));
            btnStart.setToolTipText("Iniciar");
            break;
         default:
            break;
      }
      btnStart.setEnabled(!server.isReadOnly());
   }

   private void serverConfigurationToFormFields() {
      final TerminalServerConfiguration config = server.getConfiguration();
      if (config.address().value() == null || config.address().value().isEmpty()) {
         address.setSelectedIndex(0);
      } else {
         address.getModel().setSelectedItem(config.address().value());
      }
      port.setValue(config.port().value());
      serverLogPathName.setText(config.serverLogPathName().text());
      serverLogLevel.getModel().setSelectedItem(config.serverLogLevel().value().name());
      appLogPathName.setText(config.appLogPathName().text());
      appLogLevel.getModel().setSelectedItem(config.appLogLevel().value().name());
      teLogPathName.setText(config.teLogPathName().text());
      teLogLevel.getModel().setSelectedItem(config.teLogLevel().value().name());
      configChangeCount = 0;
      btnConfirmConfig.setEnabled(false);
   }

   private void loadInterfaces() {
      final ArrayList<String> addresses = new ArrayList<>();
      addresses.add("Default");
      try {
         Enumeration<NetworkInterface> nets = NetworkInterface.getNetworkInterfaces();
         while (nets.hasMoreElements()) {
            final NetworkInterface net = nets.nextElement();
            if (net.isUp()) {
               Enumeration<InetAddress> addrs = net.getInetAddresses();
               while (addrs.hasMoreElements()) {
                  final InetAddress addr = addrs.nextElement();
                  addresses.add(addr.getHostAddress());
               }
            }
         }
      } catch (SocketException ex) {
         LOG.error("It was not possible list network interfaces", ex);
      }
      address.setModel(new DefaultComboBoxModel<>(addresses.toArray(String[]::new)));
   }

   @Override
   public void configurationChange(String name, Object oldValue, Object newValue) {
      if (server.isEmbedded() && configChangeCount == 0 && !isValueEquals(oldValue, newValue)) {
         serverConfigurationToFormFields();
      }
   }

   private boolean isValueEquals(Object oldValue, Object newValue) {
      if (oldValue == null) {
         return newValue == null;
      }
      return oldValue.equals(newValue);
   }

   private void loadPreferences() {
      final Dimension tela = Toolkit.getDefaultToolkit().getScreenSize();
      final Properties prefs = PropertiesUtil.loadPropertiesFromFile(PREFERENCES_FILE);
      int width = PropertiesUtil.getProperty(prefs, "mainWindow.width", 498);
      int height = PropertiesUtil.getProperty(prefs, "mainWindow.height", 375);
      int row = PropertiesUtil.getProperty(prefs, "mainWindow.row", (tela.height - height) / 2);
      int col = PropertiesUtil.getProperty(prefs, "mainWindow.col", (tela.width - width) / 2);
      setSize(width, height);
      setLocation(col, row);
   }

   private void savePreferences() {
      final Properties prefs = new Properties();
      prefs.setProperty("mainWindow.width", Integer.toString(getWidth()));
      prefs.setProperty("mainWindow.height", Integer.toString(getHeight()));
      prefs.setProperty("mainWindow.row", Integer.toString(getLocation().y));
      prefs.setProperty("mainWindow.col", Integer.toString(getLocation().x));
      PropertiesUtil.savePropertiesToFile(prefs, PREFERENCES_FILE, false);
   }

   @Override
   public void statusChanged(final TerminalServer server, final ServerStatus previousStatus) {
      final ServerStatus serverStatus = server.getStatus();
      switch (serverStatus) {
         case RUNNING:
            configureComponents(server);
            btnStart.setIcon(new javax.swing.ImageIcon(getClass().getResource("/stop.png")));
            btnStart.setToolTipText(TerminalUtil.getMessage("ServerMainWindow.btnStop.toolTipText"));
            status.setText(TerminalUtil.getMessage("Server.status.running"));
            break;
         case STOPPED:
            btnStart.setIcon(new javax.swing.ImageIcon(getClass().getResource("/start.png")));
            btnStart.setToolTipText(TerminalUtil.getMessage("ServerMainWindow.btnStart.toolTipText"));
            status.setText(TerminalUtil.getMessage("Server.status.stopped"));
            configureComponents(server);
            break;
         default:
            break;
      }
   }

   @Override
   public void notification(TerminalServer server, String message) {
      LOG.info(message);
   }

   @Override
   public void configurationLoaded(TerminalServer server) {
      // ignore
   }

   @Override
   public void configurationSaved(TerminalServer server) {
      if (!server.isEmbedded()) {
         serverConfigurationToFormFields();
      }
   }

   @Override
   public void sessionOpen(Session session) {
      if (server != null) {
         java.awt.EventQueue.invokeLater(() -> sessionsCount.setText(Integer.toString(server.getSessionsCount())));
      }
   }

   @Override
   public void sessionClose(Session session) {
      if (server != null) {
         java.awt.EventQueue.invokeLater(() -> sessionsCount.setText(Integer.toString(server.getSessionsCount())));
      }
   }

   @Override
   public void stateUpdate(TerminalServer server) {
      if (server != null) {
         java.awt.EventQueue.invokeLater(() -> sessionsCount.setText(Integer.toString(server.getSessionsCount())));
      }
   }

   @Override
   public void serviceStart(TerminalServer server) {
      // ignore
   }

   @Override
   public void serviceStop(TerminalServer server) {
      if (server != null) {
         java.awt.EventQueue.invokeLater(() -> sessionsCount.setText(Integer.toString(server.getSessionsCount())));
      }
   }

   private class EditionListener implements DocumentListener, ItemListener, ChangeListener {

      public EditionListener() {
         // ignore
      }

      private void configChange() {
         ++configChangeCount;
         btnConfirmConfig.setEnabled(true);
      }

      @Override
      public void insertUpdate(DocumentEvent e) {
         configChange();
      }

      @Override
      public void removeUpdate(DocumentEvent e) {
         configChange();
      }

      @Override
      public void changedUpdate(DocumentEvent e) {
         configChange();
      }

      @Override
      public void itemStateChanged(ItemEvent e) {
         configChange();
      }

      @Override
      public void stateChanged(ChangeEvent e) {
         configChange();
      }
   }
}
