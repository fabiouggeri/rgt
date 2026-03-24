/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package org.rgt.gui;

import java.util.Comparator;
import java.util.Objects;

/**
 *
 * @author fabio_uggeri
 */
public class ConfigOption implements Comparable<ConfigOption> {

   public enum OptionType {
      STRING, INT, BOOLEAN, LOG_LEVEL
   }

   private final String name;

   private OptionType type;

   public ConfigOption(final String name, final OptionType type) {
      this.name = name.trim();
      this.type = type;
   }

   public ConfigOption(final String name) {
      this(name, OptionType.STRING);
   }

   public OptionType getType() {
      return type;
   }

   public void setType(OptionType type) {
      this.type = type;
   }

   public String getName() {
      return name;
   }

   @Override
   public int compareTo(ConfigOption o) {
      if (o == null) {
         return 1;
      }
      return name.compareToIgnoreCase(o.name);
   }

   @Override
   public boolean equals(Object obj) {
      if (! (obj instanceof ConfigOption)) {
         return false;
      }
      return name.equalsIgnoreCase(((ConfigOption)obj).name);
   }

   @Override
   public int hashCode() {
      int hash = 5;
      hash = 79 * hash + Objects.hashCode(this.name);
      return hash;
   }

   @Override
   public String toString() {
      return name;
   }
}
