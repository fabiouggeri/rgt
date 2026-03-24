/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.gui.table;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.Dimension;
import java.awt.Rectangle;
import java.awt.Toolkit;
import java.awt.datatransfer.Clipboard;
import java.awt.datatransfer.StringSelection;
import java.awt.event.ActionEvent;
import java.awt.event.KeyAdapter;
import java.awt.event.KeyEvent;
import java.awt.event.MouseAdapter;
import java.awt.event.MouseEvent;
import java.beans.PropertyChangeEvent;
import java.util.HashMap;
import java.util.Map;
import javax.swing.AbstractAction;
import javax.swing.Action;
import javax.swing.BorderFactory;
import javax.swing.Box;
import javax.swing.BoxLayout;
import javax.swing.ImageIcon;
import javax.swing.JButton;
import javax.swing.JComponent;
import javax.swing.JPanel;
import javax.swing.JTable;
import javax.swing.JTextField;
import javax.swing.JToolBar;
import javax.swing.KeyStroke;
import javax.swing.UIManager;
import javax.swing.event.TableModelEvent;
import javax.swing.table.JTableHeader;
import javax.swing.table.TableColumn;
import javax.swing.table.TableColumnModel;
import javax.swing.table.TableModel;

/**
 *
 * @author fabio_uggeri
 */
public class TableFilter {

   private final JTable table;
   private boolean sortable = true;
   private boolean filterable = true;

   public static TableFilter forTable(final JTable table) {
      return new TableFilter(table);
   }

   public static void limpaFiltros(JTable table) {
      if (table.getRowSorter() instanceof EditableTableRowSorter) {
         EditableTableRowSorter rowSorter = (EditableTableRowSorter)table.getRowSorter();
         if (rowSorter.getRowFilter() instanceof FilterPopup) {
            ((FilterPopup)rowSorter.getRowFilter()).limpaFiltros();
            table.getTableHeader().repaint();
         }
      }
   }

   private TableFilter(JTable table) {
      this.table = table;
   }

   public TableFilter sortable(boolean sortable) {
      this.sortable = sortable;
      return this;
   }

   public TableFilter filterable(boolean filterable) {
      this.filterable = filterable;
      return this;
   }

   public JTable apply() {
      final KeyStroke ks;
      final Action copyAction;
      if (sortable || filterable) {
         configureRowSorter();
      }
      table.addPropertyChangeListener("tableHeader", (PropertyChangeEvent evt) -> {
         final JTable tableEvt = (JTable) evt.getSource();
         if (tableEvt.getRowSorter() instanceof EditableTableRowSorter) {
            setupTableHeader((JTableHeader) evt.getNewValue(), (EditableTableRowSorter) tableEvt.getRowSorter());
         }
      });
      copyAction = new CopyAction("copyAction");
      ks = KeyStroke.getKeyStroke("control C");
      table.getActionMap().put("copyAction", copyAction);
      table.getInputMap(JComponent.WHEN_ANCESTOR_OF_FOCUSED_COMPONENT).put(ks, "copyAction");
      return table;
   }

   private void configureRowSorter() {
      final EditableTableRowSorter rowSorter = new EditableTableRowSorter(table.getModel());
      rowSorter.setSortable(sortable);
      if (filterable) {
         configureFilter(rowSorter);
      }
      table.setRowSorter(rowSorter);
      setupTableHeader(table.getTableHeader(), rowSorter);
   }

   private void setupTableHeader(final JTableHeader header, final EditableTableRowSorter rowSorter) {
      final CustomTableHeaderRenderer renderer;
      renderer = new CustomTableHeaderRenderer(header.getDefaultRenderer());
      renderer.setRowSorter(rowSorter);
      renderer.setIcon(new ImageIcon(EditableTableRowSorter.class.getResource("/filter16.png")));
      header.setDefaultRenderer(renderer);
      header.addMouseListener(new MouseAdapter() {
         @Override
         public void mouseReleased(MouseEvent e) {
            if (e.isPopupTrigger()) {
               final JTable tableEvt = ((JTableHeader) e.getSource()).getTable();
               if (tableEvt.getRowSorter() instanceof EditableTableRowSorter) {
                  final EditableTableRowSorter rowSorter = (EditableTableRowSorter) tableEvt.getRowSorter();
                  if (rowSorter.getRowFilter() instanceof FilterPopup) {
                     ((FilterPopup) rowSorter.getRowFilter()).showFilterPopup(e);
                  }
               }
            }
            ((Component) e.getSource()).repaint();
         }
      });
   }

   private void configureFilter(final EditableTableRowSorter sorter) {
      FilterPopup filter = new FilterPopup(table);
      sorter.setRowFilter(filter);
   }

   private class FilterPopup extends PopupWindow implements ModelRowFilter {

      private static final int POPUP_WIDTH = 250;
      private static final int POPUP_HEIGHT = 65;

      private final JTable table;

      private final JTextField textField = new JTextField();

      private final Map<Integer, String> columnsFilters = new HashMap<>();

      private final Map<Integer, Dimension> popupDimensions = new HashMap<>();

      private int columnIndex = -1;

      public FilterPopup(JTable table) {
         super(true);
         this.table = table;
      }

      @Override
      protected JComponent buildContent() {
         JPanel owner = new JPanel(new BorderLayout(3, 3));
         owner.setBorder(BorderFactory.createEmptyBorder(1, 1, 1, 1));
         owner.setPreferredSize(new Dimension(POPUP_WIDTH, POPUP_HEIGHT)); // default popup size

         Box commands = new Box(BoxLayout.LINE_AXIS);

         JToolBar toolbar = new JToolBar();
         toolbar.setFloatable(false);
         toolbar.setOpaque(false);
         toolbar.add(new PopupWindow.CommandAction(
                 "Limpar filtros de todas as colunas",
                 new ImageIcon(getClass().getResource("/filter_remove16.png"))) {
            @Override
            protected boolean perform() {
               return limpaFiltros();
            }
         });

         toolbar.add(Box.createHorizontalGlue());

         toolbar.add(new JButton(new PopupWindow.CommandAction(
                 "Aplicar",
                 new ImageIcon(getClass().getResource("/confirm16.png"))) {
            @Override
            protected boolean perform() {
               return aplicaNovoFiltro();
            }
         })
         );
         toolbar.add(Box.createHorizontalStrut(5));
         toolbar.add(new JButton(new PopupWindow.CommandAction(
                 "Cancelar",
                 new ImageIcon(getClass().getResource("/cancel16.png")))));
         commands.add(toolbar);
         commands.setBorder(BorderFactory.createEmptyBorder(3, 0, 3, 0));
         commands.setBackground(UIManager.getColor("Panel.background"));
         commands.setOpaque(true);

         owner.add(textField, BorderLayout.NORTH);
         textField.addKeyListener(new KeyAdapter() {
            @Override
            public void keyPressed(KeyEvent e) {
               if (e.getKeyCode() == KeyEvent.VK_ENTER) {
                  aplicaNovoFiltro();
                  hide();
                  e.consume();
               } else if (e.getKeyCode() == KeyEvent.VK_ESCAPE) {
                  hide();
                  e.consume();
               }
            }
         });
         owner.add(commands, BorderLayout.SOUTH);
         return owner;
      }

      public void showFilterPopup(MouseEvent e) {
         final JTableHeader header = (JTableHeader) (e.getSource());
         final TableColumnModel colModel = table.getColumnModel();

         // The index of the column whose header was clicked
         int vColumnIndex = colModel.getColumnIndexAtX(e.getX());
         if (vColumnIndex < 0) {
            return;
         }

         // Determine if mouse was clicked between column heads
         Rectangle headerRect = table.getTableHeader().getHeaderRect(vColumnIndex);
         if (vColumnIndex == 0) {
            headerRect.width -= 2;
         } else {
            headerRect.grow(-2, 0);
         }

         // Mouse was clicked between column heads
         if (!headerRect.contains(e.getX(), e.getY())) {
            return;
         }

         // restore popup's size for the column
         columnIndex = table.convertColumnIndexToModel(vColumnIndex);
         setPreferredSize(getDimension(colModel.getColumn(columnIndex)));
         String text = columnsFilters.get(columnIndex);
         if (text != null) {
            textField.setText(text);
         } else {
            textField.setText("");
         }
         // show pop-up
         show(header, headerRect.x, header.getHeight());
         textField.requestFocus();
      }

      private boolean aplicaNovoFiltro() {
         if (columnIndex >= 0) {
            final String novoFiltro = textField.getText().trim();
            if (!novoFiltro.isEmpty()) {
               columnsFilters.put(columnIndex, novoFiltro.toLowerCase());
            } else {
               columnsFilters.remove(columnIndex);
            }
            table.tableChanged(new TableModelEvent(table.getModel()));
            return true;
         }
         return false;
      }

      private boolean limpaFiltros() {
         columnsFilters.clear();
         table.tableChanged(new TableModelEvent(table.getModel()));
         return true;
      }

      @Override
      public boolean include(TableModel model, int row) {
         for (Map.Entry<Integer, String> entry : columnsFilters.entrySet()) {
            final Object value = model.getValueAt(row, entry.getKey());
            if (value != null && !value.toString().toLowerCase().contains(entry.getValue())) {
               return false;
            }
         }
         return true;
      }

      @Override
      public boolean isFiltered(int column) {
         return columnsFilters.containsKey(column);
      }

      private Dimension getDimension(final TableColumn column) {
         Dimension dim = popupDimensions.get(columnIndex);
         if (dim == null) {
            dim = new Dimension(Math.max(column.getWidth(), 84), POPUP_HEIGHT);
            popupDimensions.put(columnIndex, dim);
         }
         return dim;
      }

      @Override
      protected void beforeHide() {
         if (columnIndex >= 0) {
            popupDimensions.put(columnIndex, getPreferredSize());
         }
         table.getTableHeader().repaint();
      }


   }

   private class CopyAction extends AbstractAction {

      public CopyAction(String string) {
         super(string);
      }

      @Override
      public void actionPerformed(ActionEvent evt) {
         final int col = table.getSelectedColumn();
         final int row = table.getSelectedRow();
         final Object value = table.getValueAt(row, col);
         if (value != null) {
            final Clipboard cb = Toolkit.getDefaultToolkit().getSystemClipboard();
            final StringSelection text = new StringSelection(value.toString());
            cb.setContents(text, text);
         }
      }
   }
}
