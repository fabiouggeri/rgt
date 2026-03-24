/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.gui.table;

import javax.swing.table.TableModel;

/**
 *
 * @author fabio_uggeri
 * @param <T>
 */
public interface ModelRowFilter<T extends TableModel> {

   public boolean include(final T model, final int row);

   public boolean isFiltered(final int column);
}
