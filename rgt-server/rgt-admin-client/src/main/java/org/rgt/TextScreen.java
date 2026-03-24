/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt;

import java.awt.Color;

/**
 *
 * @author fabio_uggeri
 */
public class TextScreen {

   private static final int BRIGHT_ALPHA = 128;
   private static final int DARK_ALPHA = 255;


   private static final Color[] COLORS = new Color[]{Color.BLACK,
                                                     Color.BLUE,
                                                     new Color(0, 170, 0, DARK_ALPHA),
                                                     new Color(0, 170, 255, DARK_ALPHA),
                                                     Color.RED,
                                                     Color.MAGENTA,
                                                     new Color(166, 145, 92, BRIGHT_ALPHA),
                                                     Color.LIGHT_GRAY,
                                                     Color.GRAY,
                                                     new Color(0, 0, 255, BRIGHT_ALPHA),
                                                     new Color(0, 255, 0, BRIGHT_ALPHA),
                                                     new Color(0, 255, 255, BRIGHT_ALPHA),
                                                     new Color(255, 0, 0, BRIGHT_ALPHA),
                                                     new Color(255, 0, 255, BRIGHT_ALPHA),
                                                     Color.YELLOW,
                                                     Color.WHITE};

   private static final int FOREGROUND = 0x0F;

   private static final int BACKGROUND = 0xF0;

   private final int rows;

   private final int columns;

   private int cursorRow = 0;

   private int cursorCol = 0;

   private int cursorStyle = 0;

   private final char text[];

   private final byte colors[];

   public TextScreen(int rows, int columns) {
      this.rows = rows;
      this.columns = columns;
      this.text = new char[columns * rows];
      this.colors = new byte[columns * rows];
   }

   public int getCursorStyle() {
      return cursorStyle;
   }

   public void setCursorStyle(int cursorStyle) {
      this.cursorStyle = cursorStyle;
   }

   /**
    * @return the cursorRow
    */
   public int getCursorRow() {
      return cursorRow;
   }

   /**
    * @param cursorRow the cursorRow to set
    */
   public void setCursorRow(int cursorRow) {
      this.cursorRow = cursorRow;
   }

   /**
    * @return the cursorCol
    */
   public int getCursorCol() {
      return cursorCol;
   }

   /**
    * @param cursorCol the cursorCol to set
    */
   public void setCursorCol(int cursorCol) {
      this.cursorCol = cursorCol;
   }

   /**
    * @return the rows
    */
   public int getRows() {
      return rows;
   }

   /**
    * @return the columns
    */
   public int getColumns() {
      return columns;
   }

   private int cellIndex(final int row, final int col) {
      return row * columns + col;
   }

   public void put(final int row, final int col, final char character, final byte color) {
      final int index;
      if (! isValidPos(row, col)) {
         return;
      }
      index = cellIndex(row, col);
      text[index] = character;
      colors[index] = color;
   }

   private boolean isValidPos(final int row, final int col) {
      return row >= 0 && row < rows && col >= 0 && col < columns;
   }

   public void put(final int row, final int col, final char character) {
      if (! isValidPos(row, col)) {
         return;
      }
      text[cellIndex(row, col)] = character;
   }

   public void put(final int row, final int col, final byte color) {
      if (! isValidPos(row, col)) {
         return;
      }
      colors[cellIndex(row, col)] = color;
   }

   public char getChar(final int row, final int col) {
      if (! isValidPos(row, col)) {
         return '\0';
      }
      return text[cellIndex(row, col)];
   }

   public Color getBGColor(final int row, final int col) {
      if (! isValidPos(row, col)) {
         return Color.BLACK;
      }
      return toColor(colors[cellIndex(row, col)], BACKGROUND);
   }

   public Color getFGColor(final int row, final int col) {
      if (! isValidPos(row, col)) {
         return Color.WHITE;
      }
      return toColor(colors[cellIndex(row, col)], FOREGROUND);
   }

   private Color toColor(byte color, int mask) {
      int colorIndex = color & mask;

      if (mask == BACKGROUND) {
         colorIndex >>= 4;
      }
      return COLORS[colorIndex];
   }
}
