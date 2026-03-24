/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.gui.table;

import org.rgt.client.RemoteTerminalServer;
import java.io.Serializable;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Calendar;
import java.util.Collections;
import java.util.Iterator;
import java.util.List;
import javax.swing.event.TableModelEvent;
import javax.swing.event.TableModelListener;
import javax.swing.table.TableModel;
import org.rgt.TerminalUtil;

/**
 *
 * @author fabio_uggeri
 */
public class ServerTableModel implements TableModel, Serializable {

   private static final long serialVersionUID = 8124419737163631163L;

   private final SimpleDateFormat DATE_FORMATTER = new SimpleDateFormat("dd/MM/yyyy HH:mm:ss");

   private final ArrayList<TableModelListener> listeners = new ArrayList<>();

   private final ArrayList<RemoteTerminalServer> servers = new ArrayList<>();

   public ServerTableModel() {
   }

   @Override
   public int getRowCount() {
      return servers.size();
   }

   @Override
   public int getColumnCount() {
      return 6;
   }

   @Override
   public String getColumnName(int columnIndex) {
      switch (columnIndex) {
         case 0:
            return TerminalUtil.getMessage("AdminClientWindow.servers.col.id");
         case 1:
            return TerminalUtil.getMessage("AdminClientWindow.servers.col.addr");
         case 2:
            return TerminalUtil.getMessage("AdminClientWindow.servers.col.status");
         case 3:
            return TerminalUtil.getMessage("AdminClientWindow.servers.col.sessions");
         case 4:
            return TerminalUtil.getMessage("AdminClientWindow.servers.col.startup");
         case 5:
            return TerminalUtil.getMessage("AdminClientWindow.servers.col.version");
         default:
            return "";
      }
   }

   @Override
   public Class<?> getColumnClass(int columnIndex) {
      switch (columnIndex) {
         case 0:
            return String.class;
         case 1:
            return String.class;
         case 2:
            return String.class;
         case 3:
            return Integer.class;
         case 4:
            return String.class;
         case 5:
            return String.class;
         default:
            return Object.class;
      }
   }

   @Override
   public boolean isCellEditable(int rowIndex, int columnIndex) {
      return false;
   }

   @Override
   public Object getValueAt(int rowIndex, int columnIndex) {
      if (rowIndex >= 0 && rowIndex < servers.size()) {
         final RemoteTerminalServer server = servers.get(rowIndex);
         switch (columnIndex) {
            case 0:
               return server.getId();
            case 1:
               return server.getRemoteServerAddress();
            case 2:
               return server.getStatus().toString();
            case 3:
               return server.getSessionsCount();
            case 4:
               return server.getStartTime() > 0 ? formatTimeMillis(server.getStartTime()) : "";
            case 5:
               return server.getVersion();
            default:
               return null;
         }
      }
      return null;
   }

   @Override
   public void setValueAt(Object aValue, int rowIndex, int columnIndex) {
      // Do not allow edition
   }

   private void notifyRowChange(int row, int event) {
      fireTableChanged(new TableModelEvent(this, row, row, TableModelEvent.ALL_COLUMNS, event));
   }

   public void addServer(RemoteTerminalServer server) {
      int row = servers.size();
      servers.add(server);
      notifyRowChange(row, TableModelEvent.INSERT);
   }

   public void addServers(List<RemoteTerminalServer> servers) {
      servers.addAll(servers);
      fireTableDataChanged();
   }

   public void removeServer(int row) {
      servers.remove(row);
      notifyRowChange(row, TableModelEvent.DELETE);
   }

   public void removeServer(RemoteTerminalServer server) {
      final Iterator<RemoteTerminalServer> it = servers.iterator();
      int row = 0;
      while (it.hasNext()) {
         final RemoteTerminalServer s = it.next();
         if (s.equals(server)) {
            it.remove();
            notifyRowChange(row, TableModelEvent.DELETE);
            return;
         }
         ++row;
      }
   }

   public void clearServers() {
      servers.clear();
      fireTableDataChanged();
   }

   public RemoteTerminalServer getServerAt(int index) {
      return servers.get(index);
   }

   public int indexOfServer(final String serverId) {
      for (int i = 0; i < servers.size(); i++) {
         if (servers.get(i).getId().equalsIgnoreCase(serverId)) {
            return i;
         }
      }
      return -1;
   }

   public int viewIndexOfServer(final String serverId) {
      for (int i = 0; i < servers.size(); i++) {
         if (servers.get(i).getId().equalsIgnoreCase(serverId)) {
            return i;
         }
      }
      return -1;
   }

   public RemoteTerminalServer findServer(final String serverId) {
      final int index = indexOfServer(serverId);
      if (index >= 0) {
         return servers.get(index);
      }
      return null;
   }

   private String formatTimeMillis(long startTime) {
      final Calendar calendar = Calendar.getInstance();
      calendar.setTimeInMillis(startTime);
      DATE_FORMATTER.setCalendar(calendar);
      return DATE_FORMATTER.format(calendar.getTime());
   }

   @Override
   public void addTableModelListener(TableModelListener l) {
      listeners.add(l);
   }

   @Override
   public void removeTableModelListener(TableModelListener listener) {
      listeners.remove(listener);
   }

   public TableModelListener[] getTableModelListeners() {
      return listeners.toArray(new TableModelListener[listeners.size()]);
   }

   public void fireTableDataChanged() {
      fireTableChanged(new TableModelEvent(this));
   }

   public void fireTableChanged(TableModelEvent evt) {
      listeners.forEach(listener -> listener.tableChanged(evt));
   }

   public List<RemoteTerminalServer> getServers() {
      return Collections.unmodifiableList(servers);
   }

   public void notifyServerUpdate(int row) {
      if (row >= 0 && row < servers.size()) {
         notifyRowChange(row, TableModelEvent.UPDATE);
      }
   }
}
