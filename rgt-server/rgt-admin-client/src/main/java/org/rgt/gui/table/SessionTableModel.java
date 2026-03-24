/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.gui.table;

import org.rgt.Session;
import java.io.Serializable;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Calendar;
import java.util.Collection;
import java.util.Iterator;
import javax.swing.event.TableModelEvent;
import javax.swing.event.TableModelListener;
import javax.swing.table.TableModel;
import org.rgt.TerminalUtil;

/**
 *
 * @author fabio_uggeri
 */
public class SessionTableModel implements TableModel, Serializable {

   private static final long serialVersionUID = 1080972125563294460L;

   /** List of listeners */
   private final ArrayList<TableModelListener> listeners = new ArrayList<>();

   private final ArrayList<Session> sessions = new ArrayList<>();

   @Override
   public int getRowCount() {
      return sessions.size();
   }

   @Override
   public int getColumnCount() {
      return 7;
   }

   @Override
   public String getColumnName(int columnIndex) {
      switch(columnIndex) {
         case 0:
            return TerminalUtil.getMessage("SessionsWindow.sessions.col.id");
         case 1:
            return TerminalUtil.getMessage("SessionsWindow.sessions.col.terminal_addr");
         case 2:
            return TerminalUtil.getMessage("SessionsWindow.sessions.col.os_user");
         case 3:
            return TerminalUtil.getMessage("SessionsWindow.sessions.col.pid_app");
         case 4:
            return TerminalUtil.getMessage("SessionsWindow.sessions.col.status");
         case 5:
            return TerminalUtil.getMessage("SessionsWindow.sessions.col.connection_time");
         case 6:
            return TerminalUtil.getMessage("SessionsWindow.sessions.col.cmd_line");
         default:
            return "";
      }
   }

   @Override
   public Class<?> getColumnClass(int columnIndex) {
      switch(columnIndex) {
         case 0:
            return Long.class;
         case 1:
            return String.class;
         case 2:
            return String.class;
         case 3:
            return Long.class;
         case 4:
            return String.class;
         case 5:
            return String.class;
         case 6:
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
      if (rowIndex >= 0 && rowIndex < sessions.size()) {
         switch(columnIndex) {
            case 0:
               return sessions.get(rowIndex).getId();
            case 1:
               return sessions.get(rowIndex).getTerminalAddress();
            case 2:
               return sessions.get(rowIndex).getOSUser();
            case 3:
               return sessions.get(rowIndex).getAppPid();
            case 4:
               return sessions.get(rowIndex).getStatus().toString();
            case 5:
               return formatTimeMillis(sessions.get(rowIndex).getStartTime());
            case 6:
               return sessions.get(rowIndex).getCommandLine();
            default:
               return null;
         }
      }
      return null;
   }

   public Session getSession(int rowIndex) {
      if (rowIndex >= 0 && rowIndex < sessions.size()) {
         return sessions.get(rowIndex);
      }
      return null;
   }

   @Override
   public void setValueAt(Object aValue, int rowIndex, int columnIndex) {
      // Not allow edition
   }

   public void addSession(Session session) {
      int row = sessions.size();
      sessions.add(session);
      fireTableChanged(new TableModelEvent(this, row, row,
             TableModelEvent.ALL_COLUMNS, TableModelEvent.INSERT));
   }

   public void addSessions(Collection<Session> sessions) {
      addSessions(sessions, true);
   }

   public void addSessions(Collection<Session> sessions, final boolean notifyDataChanged) {
      this.sessions.addAll(sessions);
      if (notifyDataChanged) {
         fireTableDataChanged();
      }
   }

   public void removeSession(Session session) {
      final Iterator<Session> it = sessions.iterator();
      int row = 0;
      while (it.hasNext()) {
         final Session s = it.next();
         if (s.getId() == session.getId()) {
            it.remove();
            fireTableChanged(new TableModelEvent(this, row, row,
                   TableModelEvent.ALL_COLUMNS, TableModelEvent.DELETE));
            return;
         }
         ++row;
      }
   }

   public void removeSession(long sessionId) {
      final Iterator<Session> it = sessions.iterator();
      int row = 0;
      while (it.hasNext()) {
         final Session s = it.next();
         if (s.getId() == sessionId) {
            it.remove();
            fireTableChanged(new TableModelEvent(this, row, row,
                   TableModelEvent.ALL_COLUMNS, TableModelEvent.DELETE));
            return;
         }
         ++row;
      }
   }

   public void clearSessions() {
      clearSessions(true);
   }

   public void clearSessions(boolean notifyDataChanged) {
      sessions.clear();
      if (notifyDataChanged) {
         fireTableDataChanged();
      }
   }

   private String formatTimeMillis(long startTime) {
      final Calendar c = Calendar.getInstance();
      final SimpleDateFormat df = new SimpleDateFormat("dd/MM/yyyy HH:mm:ss");
      c.setTimeInMillis(startTime);
      df.setCalendar(c);
      return df.format(c.getTime());
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
}
