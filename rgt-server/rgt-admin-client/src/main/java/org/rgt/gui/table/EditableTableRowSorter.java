/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.gui.table;

import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.List;
import javax.swing.RowSorter;
import javax.swing.SortOrder;
import javax.swing.table.TableModel;

/**
 *
 * @author fabio_uggeri
 * @param <T>
 */
public class EditableTableRowSorter<T extends TableModel> extends RowSorter<T> {

   private T model;

   private boolean multiColumnSort = false;

   private boolean multiColumnFilter = true;

   private boolean sortsOnUpdates = false;

   private boolean sortsOnInsert = false;

   private boolean filterOnUpdates = false;

   private boolean filterOnInsert = false;

   private final ArrayList<RowSorter.SortKey> sortKeys = new ArrayList<>();

   private ModelRowFilter<T> rowFilter = null;

   private int modelRowCount = 0;

   private final ArrayList<Integer> viewToModel = new ArrayList<>();

   private final ArrayList<Integer> modelToView = new ArrayList<>();

   private boolean[] columnSortable = null;

   private boolean sortable = true;

   private boolean sorted = false;

   private final Comparator<Integer> sortComparator = new ColumnsComparator();

   private RowComparator rowComparator = null;

   private Comparator[] comparators = null;

   public EditableTableRowSorter(T model) {
      this.model = model;
      this.modelRowCount = model.getRowCount();
   }

   public boolean isSortsOnUpdates() {
      return sortsOnUpdates;
   }

   public boolean isSortsOnInsert() {
      return sortsOnInsert;
   }

   public boolean isMultiColumnSort() {
      return multiColumnSort;
   }

   public boolean isMultiColumnFilter() {
      return multiColumnFilter;
   }

   public boolean isFilterOnUpdates() {
      return filterOnUpdates;
   }

   public boolean isFilterOnInsert() {
      return filterOnInsert;
   }

   public void setSortsOnUpdates(boolean sortsOnUpdates) {
      this.sortsOnUpdates = sortsOnUpdates;
   }

   public void setSortsOnInsert(boolean sortsOnInsert) {
      this.sortsOnInsert = sortsOnInsert;
   }

   public void setMultiColumnSort(boolean multiColumnSort) {
      this.multiColumnSort = multiColumnSort;
   }

   public void setMultiColumnFilter(boolean multiColumnFilter) {
      this.multiColumnFilter = multiColumnFilter;
   }

   public void setFilterOnInsert(boolean filterOnInsert) {
      this.filterOnInsert = filterOnInsert;
   }

   public void setFilterOnUpdates(boolean filterOnUpdates) {
      this.filterOnUpdates = filterOnUpdates;
   }

   public boolean isSortable(int column) {
      checkColumnModelIndex(column);
      return sortable && (columnSortable == null || columnSortable[column]);
   }

   public void setSortable(int column, boolean sort) {
      checkColumnModelIndex(column);
      if (columnSortable == null) {
         columnSortable = new boolean[model.getColumnCount()];
      }
      columnSortable[column] = sort;
   }

   public void setSortable(boolean sort) {
      this.sortable = sort;
   }

   @Override
   public T getModel() {
      return model;
   }

   public void setModel(T model) {
      this.model = model;
      modelStructureChanged();
   }

   private int findSortKey(int column) {
      for (int i = 0; i < sortKeys.size(); i++) {
         if (sortKeys.get(i).getColumn() == column) {
            return i;
         }
      }
      return -1;
   }

   @Override
   public void toggleSortOrder(int column) {
      final int sortKeyPos;
      checkColumnModelIndex(column);
      if (isSortable(column)) {
         final int[] lastViewToModel = intListToArray(viewToModel);
         sorted = false;
         if (multiColumnSort) {
            sortKeyPos = findSortKey(column);
            if (sortKeyPos == 0) {
               sortKeys.set(0, new RowSorter.SortKey(column, inverseOrder(sortKeys.get(0).getSortOrder())));
            } else if (sortKeyPos > 0) {
               sortKeys.remove(column);
               sortKeys.add(0, new RowSorter.SortKey(column, SortOrder.ASCENDING));
            } else {
               sortKeys.add(0, new RowSorter.SortKey(column, SortOrder.ASCENDING));
            }
         } else if (!sortKeys.isEmpty()) {
            SortOrder newOrder = inverseOrder(sortKeys.get(0).getSortOrder());
            sortKeys.clear();
            sortKeys.add(new RowSorter.SortKey(column, newOrder));
         } else {
            sortKeys.add(new RowSorter.SortKey(column, SortOrder.ASCENDING));
         }
         fireRowSorterChanged(lastViewToModel);
         sortRows();
      }
   }

   @Override
   public int convertRowIndexToModel(int index) {
      if (index < 0 || index >= viewToModel.size()) {
         return -1;
      }
      return viewToModel.get(index);
   }

   @Override
   public int convertRowIndexToView(int index) {
      if (index < 0 || index >= modelToView.size()) {
         return -1;
      }
      return modelToView.get(index);
   }

   @Override
   public void setSortKeys(List<? extends RowSorter.SortKey> keys) {
      final int[] lastViewToModel = intListToArray(viewToModel);
      sortKeys.clear();
      if (keys != null) {
         sortKeys.addAll(keys);
      }
      fireRowSorterChanged(lastViewToModel);
      sortRows();
   }

   private int[] intListToArray(List<Integer> list) {
      if (list != null) {
         final int[] intArray = new int[list.size()];
         for (int i = 0; i < intArray.length; i++) {
            intArray[i] = list.get(i);
         }
         return intArray;
      } else {
         return new int[0];
      }
   }

   private void sortRows() {
      if (!sorted) {
         this.sorted = true;
         Collections.sort(viewToModel, sortComparator);
         for (int i = 0; i < modelToView.size(); i++) {
            modelToView.set(i, -1);
         }
         for (int i = 0; i < viewToModel.size(); i++) {
            modelToView.set(viewToModel.get(i), i);
         }
      }
   }

   private void filterRows() {
      modelToView.clear();
      viewToModel.clear();
      if (rowFilter != null) {
         for (int i = 0; i < model.getRowCount(); i++) {
            if (rowFilter.include(model, i)) {
               modelToView.add(viewToModel.size());
               viewToModel.add(i);
            } else {
               modelToView.add(-1);
            }
         }
      } else {
         for (int i = 0; i < model.getRowCount(); i++) {
            modelToView.add(viewToModel.size());
            viewToModel.add(i);
         }
      }
   }

   @Override
   public List<? extends RowSorter.SortKey> getSortKeys() {
      return Collections.unmodifiableList(sortKeys);
   }

   @Override
   public int getViewRowCount() {
      return viewToModel.size();
   }

   @Override
   public int getModelRowCount() {
      return model.getRowCount();
   }

   @Override
   public void modelStructureChanged() {
      final int[] lastViewToModel = intListToArray(viewToModel);
      sortKeys.clear();
      modelRowCount = model.getRowCount();
      viewToModel.clear();
      modelToView.clear();
      columnSortable = null;
      filterRows();
      sortRows();
      fireRowSorterChanged(lastViewToModel);
   }

   @Override
   public void allRowsChanged() {
      final int[] lastViewToModel = intListToArray(viewToModel);
      modelRowCount = model.getRowCount();
      filterRows();
      sortRows();
      fireRowSorterChanged(lastViewToModel);
   }

   @Override
   public void rowsInserted(int firstRow, int endRow) {
      final int[] lastViewToModel;
      modelRowCount = model.getRowCount();
      checkRowsModelIndex(firstRow, endRow);
      lastViewToModel = intListToArray(viewToModel);
      if (filterOnInsert) {
         filterRows();
      } else {
         final int initialViewSize = viewToModel.size();
         final int delta = (endRow - firstRow) + 1;
         for (int i = firstRow; i <= endRow; i++) {
            modelToView.add(i, viewToModel.size());
            viewToModel.add(i);
         }
         for (int i = 0; i < initialViewSize; i++) {
            final int modelIndex = viewToModel.get(i);
            if (modelIndex >= firstRow) {
               viewToModel.set(i, modelIndex + delta);
            }
         }
      }
      this.sorted = false;
      if (sortsOnInsert) {
         sortRows();
      }
      fireRowSorterChanged(lastViewToModel);
   }

   @Override
   public void rowsDeleted(int firstRow, int endRow) {
      final int[] lastViewToModel;
      checkRowsModelIndex(firstRow, endRow);
      lastViewToModel = intListToArray(viewToModel);
      if (modelToView != null && viewToModel != null) {
         final int delta = (endRow - firstRow) + 1;
         for (int i = endRow; i >= firstRow; i--) {
            final int viewIndex = modelToView.remove(i);
            if (viewIndex >= 0) {
               viewToModel.remove(viewIndex);
            }
         }
         for (int i = 0; i < viewToModel.size(); i++) {
            final int modelIndex = viewToModel.get(i);
            if (modelIndex > endRow) {
               viewToModel.set(i, modelIndex - delta);
            }
         }
         for (int i = 0; i < modelToView.size(); i++) {
            modelToView.set(i, -1);
         }
         for (int i = 0; i < viewToModel.size(); i++) {
            modelToView.set(viewToModel.get(i), i);
         }
      }
      modelRowCount = model.getRowCount();
      fireRowSorterChanged(lastViewToModel);
   }

   @Override
   public void rowsUpdated(int firstRow, int endRow) {
      final int[] lastViewToModel;
      checkRowsModelIndex(firstRow, endRow);
      lastViewToModel = intListToArray(viewToModel);
      if (filterOnUpdates) {
         filterRows();
      }
      this.sorted = false;
      if (sortsOnUpdates) {
         sortRows();
      }
      fireRowSorterChanged(lastViewToModel);
   }

   @Override
   public void rowsUpdated(int firstRow, int endRow, int column) {
      checkColumnModelIndex(column);
      rowsUpdated(firstRow, endRow);
   }

   private void checkRowsModelIndex(int firstRow, int endRow) {
      if (firstRow > endRow || firstRow < 0 || endRow < 0
              || firstRow > modelRowCount) {
         throw new IndexOutOfBoundsException("Invalid range: " + firstRow + ":" + endRow + ". Rows: " + modelRowCount);
      }
   }

   private void checkColumnModelIndex(int column) {
      if (column < 0 || column >= model.getColumnCount()) {
         throw new IndexOutOfBoundsException("Coluna fora dos limites da tabela: " + column);
      }
   }

   private SortOrder inverseOrder(SortOrder sortOrder) {
      if (SortOrder.ASCENDING.equals(sortOrder)) {
         return SortOrder.DESCENDING;
      } else {
         return SortOrder.ASCENDING;
      }
   }

   public void setRowFilter(ModelRowFilter<T> rowFilter) {
      this.rowFilter = rowFilter;
   }

   public ModelRowFilter<T> getRowFilter() {
      return rowFilter;
   }

   public boolean isSorted() {
      return sorted;
   }

   public Comparator<Integer> getComparator(int column) {
      checkColumnModelIndex(column);
      if (comparators != null) {
         return comparators[column];
      }
      return null;
   }

   public void setComparator(int column, Comparator<?> comparator) {
      checkColumnModelIndex(column);
      if (comparators == null) {
         comparators = new Comparator<?>[model.getColumnCount()];
      }
      comparators[column] = comparator;
   }

   public void setRowComparator(RowComparator rowComparator) {
      this.rowComparator = rowComparator;
   }

   private class ColumnsComparator implements Comparator<Integer> {

      @Override
      public int compare(Integer i1, Integer i2) {
         if (rowComparator != null) {
            int result;
            for (RowSorter.SortKey sortKey : sortKeys) {
               result = rowComparator.compare(model, sortKey, i1, i2);
               if (result != 0) {
                  return result;
               }
            }
         } else {
            for (RowSorter.SortKey sortKey : sortKeys) {
               final Comparator comp = getComparator(sortKey.getColumn());
               final Object o1 = model.getValueAt(i1, sortKey.getColumn());
               final Object o2 = model.getValueAt(i2, sortKey.getColumn());
               final int result;
               if (comp != null) {
                  result = comp.compare(o1, o2);
               } else if (o1 == null) {
                  if (o2 == null) {
                     result = 0;
                  } else {
                     result = -1;
                  }
               } else if (o2 == null) {
                  result = 1;
               } else if (o1 instanceof Comparable) {
                  result = ((Comparable) o1).compareTo(o2);
               } else {
                  result = o1.toString().compareTo(o2.toString());
               }
               if (result != 0) {
                  return SortOrder.ASCENDING.equals(sortKey.getSortOrder()) ? result : result * -1;
               }
            }
         }
         return i1 - i2;
      }
   }
}
