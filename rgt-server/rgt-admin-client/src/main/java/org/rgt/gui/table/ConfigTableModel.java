/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.gui.table;

import java.io.Serializable;
import java.util.ArrayList;
import java.util.BitSet;
import java.util.Collection;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import javax.swing.event.TableModelEvent;
import javax.swing.event.TableModelListener;
import javax.swing.table.TableModel;
import org.rgt.TerminalUtil;

/**
 *
 * @author fabio_uggeri
 */
public class ConfigTableModel implements TableModel, Serializable {

   private final ArrayList<TableModelListener> listeners = new ArrayList<>();

   private final String columns[];

   private final List<String[]> data;

   private boolean changed = false;

   private final List<BitSet> changedCells = new ArrayList<>();

   private final Set<String> removed = new HashSet<>();

   public ConfigTableModel(final String columns[], final List<String[]> data) {
      this.columns = columns;
      this.data = data;
   }

   @Override
   public int getRowCount() {
      return data.size();
   }

   @Override
   public int getColumnCount() {
      return columns.length;
   }

   @Override
   public String getColumnName(int columnIndex) {
      return columns[columnIndex];
   }

   @Override
   public Class<?> getColumnClass(int columnIndex) {
      return String.class;
   }

   @Override
   public boolean isCellEditable(int rowIndex, int columnIndex) {
      return columnIndex > 0 || TerminalUtil.isEmpty(data.get(rowIndex)[columnIndex]);
   }

   @Override
   public Object getValueAt(int rowIndex, int columnIndex) {
      Object value;
      if (rowIndex < 0 || rowIndex >= data.size()) {
         return "";
      }
      if (columnIndex < 0 || columnIndex > columns.length) {
         return "";
      }
      value = data.get(rowIndex)[columnIndex] ;
      return value == null ? "" : value;
   }

   @Override
   public void setValueAt(Object aValue, int rowIndex, int columnIndex) {
      if (rowIndex < 0 || rowIndex >= data.size()) {
         return;
      }
      if (columnIndex < 0 || columnIndex > columns.length) {
         return;
      }
      if (! data.get(rowIndex)[columnIndex].equals(aValue.toString())) {
         data.get(rowIndex)[columnIndex] = aValue.toString();
         notifyCellChange(rowIndex, columnIndex, TableModelEvent.UPDATE);
      }
   }

   @Override
   public void addTableModelListener(TableModelListener l) {
      listeners.add(l);
   }

   public void delete(int rowIndex) {
      String rowData[];
      if (rowIndex < 0 || rowIndex >= data.size()) {
         return;
      }
      rowData = data.remove(rowIndex);
      if (rowData != null && rowData.length > 0) {
         removed.add(rowData[0]);
      }
      notifyRowChange(rowIndex, TableModelEvent.DELETE);
   }

   public boolean add(final String name) {
      String newRow[];
      for (final String row[] : data) {
         if (row[0].equalsIgnoreCase(name)) {
            return false;
         }
      }
      newRow = new String[columns.length];
      newRow[0] = name;
      for (int i = 1; i < newRow.length; i++) {
         newRow[i] = "";
      }
      data.add(newRow);
      notifyRowChange(data.size() - 1, TableModelEvent.INSERT);
      return true;
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

   private void notifyRowChange(int row, int event) {
      changed = true;
      fireTableChanged(new TableModelEvent(this, row, row, TableModelEvent.ALL_COLUMNS, event));
   }

   private void notifyCellChange(int row, int col, int event) {
      registerChange(row, col);
      fireTableChanged(new TableModelEvent(this, row, row, col, event));
   }

   private void registerChange(int row, int col) {
      if (col <= 0) {
         return;
      }
      while (row >= changedCells.size()) {
         changedCells.add(new BitSet(columns.length));
      }
      changedCells.get(row).set(col);
      changed = true;
   }

   public boolean isChanged(final int row, final int col) {
      if (row >= 0 && row < changedCells.size() && col >= 0 && col < columns.length) {
         return changedCells.get(row).get(col);
      }
      return false;
   }

   public boolean isChanged(final int row) {
      if (row >= 0 && row < changedCells.size()) {
         return ! changedCells.get(row).isEmpty();
      }
      return false;
   }

   public boolean isChanged() {
      return changed;
   }

   public void clearChanges() {
      changed = false;
      removed.clear();
      changedCells.forEach(bs -> bs.clear());
      fireTableDataChanged();
   }

   public Collection<String> removed() {
      return removed;
   }

   public String getOptionName(final int row) {
      return getValueAt(row, 0).toString();
   }
}
