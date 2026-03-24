/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.gui.table;

import java.util.ArrayList;
import java.util.Collection;
import java.util.Comparator;
import java.util.Date;
import java.util.List;
import javax.swing.event.TableModelEvent;
import javax.swing.event.TableModelListener;
import javax.swing.table.TableModel;
import org.rgt.TerminalUtil;
import org.rgt.protocol.admin.files.FileInfo;

/**
 *
 * @author fabio_uggeri
 */
public class FileInfoModel implements TableModel {

   private final ArrayList<TableModelListener> listeners = new ArrayList<>();

   private final ArrayList<FileInfo> files = new ArrayList<>();

   public FileInfoModel() {
   }

   public FileInfoModel(List<FileInfo> files) {
      this.files.addAll(files);
   }

   @Override
   public int getRowCount() {
      return files.size();
   }

   @Override
   public int getColumnCount() {
      return 4;
   }

   @Override
   public String getColumnName(int col) {
      switch (col) {
         case 0:
            return TerminalUtil.getMessage("RemoteFilesWindow.files.col.name");
         case 1:
            return TerminalUtil.getMessage("RemoteFilesWindow.files.col.length");
         case 2:
            return TerminalUtil.getMessage("RemoteFilesWindow.files.col.creation");
         case 3:
            return TerminalUtil.getMessage("RemoteFilesWindow.files.col.modification");
         default:
            return "";
      }
   }

   @Override
   public Class<?> getColumnClass(int col) {
      switch (col) {
         case 0:
            return FileInfo.class;
         case 1:
            return Long.class;
         case 2:
            return Date.class;
         case 3:
            return Date.class;
         default:
            return Object.class;
      }
   }

   @Override
   public boolean isCellEditable(int rowIndex, int columnIndex) {
      return false;
   }

   @Override
   public Object getValueAt(int row, int col) {
      if (row >= 0 && row < files.size()) {
         final FileInfo file = files.get(row);
         switch (col) {
            case 0:
               return file;
            case 1:
               return file.getLength();
            case 2:
               return file.getCreationTime();
            case 3:
               return file.getLastModificationTime();
            default:
               return null;
         }
      }
      return null;
   }

   @Override
   public void setValueAt(Object value, int row, int col) {
   }

   public boolean addFile(final FileInfo fileInfo) {
      if (files.add(fileInfo)) {
         notifyRowChange(files.size() - 1, TableModelEvent.INSERT);
         return true;
      }
      return false;
   }

   public boolean addFiles(final Collection<FileInfo> filesInfo) {
      return addFiles(filesInfo, true);
   }

   public boolean addFiles(final Collection<FileInfo> filesInfo, final boolean notifyDataChanged) {
      if (files.addAll(filesInfo)) {
         if (notifyDataChanged) {
            fireTableDataChanged();
         }
         return true;
      }
      return false;
   }

   public FileInfo findFile(final String filename) {
      final int index = indexOf(filename);
      if (index >= 0) {
         return files.get(index);
      }
      return null;
   }

   public int indexOf(final String filename) {
      for (int i = 0; i < files.size(); i++) {
         if (files.get(i).getName().equalsIgnoreCase(filename)) {
            return i;
         }
      }
      return -1;
   }

   public int indexOf(FileInfo fileInfo) {
      for (int i = 0; i < files.size(); i++) {
         if (files.get(i).equals(fileInfo)) {
            return i;
         }
      }
      return -1;
   }

   public FileInfo fileInfoAt(int row) {
      if (row >= 0 && row < files.size()) {
         return files.get(row);
      }
      return null;
   }

   public boolean removeFile(final FileInfo fileInfo) {
      final int index = indexOf(fileInfo);
      if (index >= 0 && files.remove(index) != null) {
         notifyRowChange(index, TableModelEvent.DELETE);
         return true;
      }
      return false;
   }

   public boolean removeFile(final String filename) {
      final int index = indexOf(filename);
      if (index >= 0 && files.remove(index) != null) {
         notifyRowChange(index, TableModelEvent.DELETE);
         return true;
      }
      return false;
   }

   public void clearFiles() {
      clearFiles(true);
   }

   public void clearFiles(boolean notifyChange) {
      files.clear();
      if (notifyChange) {
         fireTableDataChanged();
      }
   }

   public void sort(Comparator<FileInfo> comparator) {
      files.sort(comparator);
   }

   private void notifyRowChange(int row, int event) {
      fireTableChanged(new TableModelEvent(this, row, row, TableModelEvent.ALL_COLUMNS, event));
   }

   @Override
   public void addTableModelListener(TableModelListener l) {
      listeners.add(l);
   }

   @Override
   public void removeTableModelListener(TableModelListener l) {
      listeners.remove(l);
   }

   public void fireTableDataChanged() {
      fireTableChanged(new TableModelEvent(this));
   }

   public void fireTableChanged(TableModelEvent evt) {
      listeners.forEach(listener -> listener.tableChanged(evt));
   }
}
