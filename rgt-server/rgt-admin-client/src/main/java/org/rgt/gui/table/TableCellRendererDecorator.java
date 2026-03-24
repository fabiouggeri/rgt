/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.gui.table;

import java.awt.Component;
import javax.swing.JTable;
import javax.swing.table.TableModel;

/**
 *
 * @author fabio_uggeri
 * @param <M>
 * @param <R>
 */
@FunctionalInterface
public interface TableCellRendererDecorator<M extends TableModel, R extends Component> {
   void decorate(final JTable table, final M model, final R renderer, final int row, final int col, boolean isSelected, boolean hasFocus);
}
