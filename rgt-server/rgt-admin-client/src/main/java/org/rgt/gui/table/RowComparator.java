/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.gui.table;

import javax.swing.RowSorter;
import javax.swing.table.TableModel;

/**
 *
 * @author fabio_uggeri
 * @param <T>
 */
@FunctionalInterface
public interface RowComparator<T extends TableModel> {

   int compare(T model, RowSorter.SortKey sortKey, int firstRow, int secondRow);

}
