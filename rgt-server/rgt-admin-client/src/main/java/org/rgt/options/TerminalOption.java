/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.options;

import java.io.File;
import java.io.IOException;
import java.io.StringReader;
import java.io.StringWriter;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Properties;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import org.rgt.RGTLogLevel;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 *
 * @author fabio_uggeri
 * @param <T> Option type
 */
public abstract class TerminalOption<T> {

   private static final Logger LOG = LoggerFactory.getLogger(TerminalOption.class);

   private static final Pattern ENV_VAR_PATTERN = Pattern.compile("%([^%]+)%");

   protected final String name;
   protected final String defaultText;
   protected String text;
   protected T value;
   protected final List<TerminalOptionListener> listeners = new ArrayList<>();

   private TerminalOption(final String name, final String defaultValue) {
      this.name = name;
      this.text = defaultValue;
      this.defaultText = defaultValue;
      this.value = resolve(defaultValue);
   }

   public static TerminalOption<String> stringType(final String name, final String defaultValue) {
      return new StringTerminalOption(name, defaultValue);
   }

   public static TerminalOption<Integer> intType(final String name, final String defaultValue) {
      return new IntTerminalOption(name, defaultValue);
   }

   public static TerminalOption<File> fileType(final String name, final String defaultValue) {
      return new FileTerminalOption(name, defaultValue);
   }

   public static TerminalOption<RGTLogLevel> logLevelType(final String name, final String defaultValue) {
      return new LogLevelTerminalOption(name, defaultValue);
   }

   public static TerminalOption<Boolean> booleanType(final String name, final String defaultValue) {
      return new BooleanTerminalOption(name, defaultValue);
   }

   public static TerminalOption<Properties> propertiesType(final String name) {
      return new PropertiesTerminalOption(name);
   }

   public abstract Class<T> getType();
   public abstract TerminalOption<T> copy();
   protected abstract T stringToValue(final String value);
   protected abstract String valueToString(T value);


   private static String resolveEnvVars(final String value) {
      final StringBuilder sb = new StringBuilder(value);
      final Matcher matcher = ENV_VAR_PATTERN.matcher(value);
      while (matcher.find()) {
         final String envVarValue = System.getenv(matcher.group(1).toUpperCase());
         if (envVarValue != null) {
            sb.replace(matcher.start(), matcher.end(), envVarValue);
         } else {
            sb.replace(matcher.start(), matcher.end(), "");
         }
      }
      return sb.toString();
   }

   @Override
   public boolean equals(Object obj) {
      if (! (obj instanceof TerminalOption)) {
         return false;
      }
      return name().equalsIgnoreCase(((TerminalOption)obj).name());
   }

   @Override
   public int hashCode() {
      int hash = 7;
      hash = 11 * hash + Objects.hashCode(this.name);
      return hash;
   }

   private T resolve(final String value) {
      return stringToValue(resolveEnvVars(value));
   }

   public String name() {
      return name;
   }

   public String defaultText() {
      return defaultText;
   }

   public String text() {
      return text;
   }

   public Map<String, String> textMap() {
      return Collections.singletonMap(name, text);
   }

   public void text(String newValue) {
      final String oldValue = this.text;
      this.text = newValue == null ? defaultText : newValue;
      this.value = resolve(this.text);
      fireConfigurationChange(name, oldValue, newValue);
   }

   public void value(T newValue) {
      final String oldValue = this.text;
      this.text = newValue == null ? defaultText : valueToString(newValue);
      this.value = newValue;
      fireConfigurationChange(name, oldValue, newValue);
   }

   public T value() {
      return value;
   }

   public void addListener(TerminalOptionListener listener) {
      listeners.add(listener);
   }

   public void addListeners(Collection<TerminalOptionListener> listeners) {
      this.listeners.addAll(listeners);
   }

   public boolean removeListener(TerminalOptionListener listener) {
      return listeners.remove(listener);
   }

   public void clearListeners() {
      listeners.clear();
   }

   public List<TerminalOptionListener> getListeners() {
      return listeners;
   }

   @Override
   public String toString() {
      return text;
   }

   private void fireConfigurationChange(final String name, final Object oldValue, final Object newValue) {
      if (oldValue != newValue && (oldValue == null || !oldValue.equals(newValue))) {
         LOG.debug("{} changed from {} to {}", name, oldValue, newValue);
         listeners.forEach((l) -> l.configurationChange(name, oldValue, newValue));
      }
   }

   private static class StringTerminalOption extends TerminalOption<String> {

      public StringTerminalOption(String name, String defaultValue) {
         super(name, defaultValue);
      }

      @Override
      public Class<String> getType() {
         return String.class;
      }

      @Override
      public String stringToValue(String value) {
         return value;
      }

      @Override
      protected String valueToString(String value) {
         return value;
      }

      @Override
      public TerminalOption<String> copy() {
         final TerminalOption<String> copy = stringType(name, defaultText);
         copy.text = this.text;
         copy.value = this.value;
         return copy;
      }
   }

   private static class IntTerminalOption extends TerminalOption<Integer> {

      public IntTerminalOption(String name, String defaultValue) {
         super(name, defaultValue);
      }

      @Override
      public Class<Integer> getType() {
         return Integer.class;
      }

      @Override
      public Integer stringToValue(String value) {
         return Integer.valueOf(value);
      }

      @Override
      protected String valueToString(Integer value) {
         return value.toString();
      }

      @Override
      public TerminalOption<Integer> copy() {
         final TerminalOption<Integer> copy = intType(name, defaultText);
         copy.text = this.text;
         copy.value = this.value;
         return copy;
      }
   }

   private static class FileTerminalOption extends TerminalOption<File> {

      public FileTerminalOption(String name, String defaultValue) {
         super(name, defaultValue);
      }

      @Override
      public Class<File> getType() {
         return File.class;
      }

      @Override
      public File stringToValue(String value) {
         return new File(value);
      }

      @Override
      protected String valueToString(File value) {
         return value.getPath();
      }

      @Override
      public TerminalOption<File> copy() {
         final TerminalOption<File> copy = fileType(name, defaultText);
         copy.text = this.text;
         copy.value = this.value;
         return copy;
      }
   }

   private static class LogLevelTerminalOption extends TerminalOption<RGTLogLevel> {

      public LogLevelTerminalOption(String name, String defaultValue) {
         super(name, defaultValue);
      }

      @Override
      public Class<RGTLogLevel> getType() {
         return RGTLogLevel.class;
      }

      @Override
      public RGTLogLevel stringToValue(String value) {
         return RGTLogLevel.valueOf(value);
      }

      @Override
      protected String valueToString(RGTLogLevel value) {
         return value.name();
      }

      @Override
      public TerminalOption<RGTLogLevel> copy() {
         final TerminalOption<RGTLogLevel> copy = logLevelType(name, defaultText);
         copy.text = this.text;
         copy.value = this.value;
         return copy;
      }
   }

   private static class BooleanTerminalOption extends TerminalOption<Boolean> {

      public BooleanTerminalOption(String name, String defaultValue) {
         super(name, defaultValue);
      }

      @Override
      public Class<Boolean> getType() {
         return Boolean.class;
      }

      @Override
      public Boolean stringToValue(String value) {
         return "true".equalsIgnoreCase(value);
      }

      @Override
      protected String valueToString(Boolean value) {
         return value.toString();
      }

      @Override
      public TerminalOption<Boolean> copy() {
         final TerminalOption<Boolean> copy = booleanType(name, defaultText);
         copy.text = this.text;
         copy.value = this.value;
         return copy;
      }
   }

   private static class PropertiesTerminalOption extends TerminalOption<Properties> {

      public PropertiesTerminalOption(String name) {
         super(name, "");
      }

      @Override
      public Class<Properties> getType() {
         return Properties.class;
      }

      @Override
      public Properties stringToValue(String value) {
         final Properties properties = new Properties();
         try {
            properties.load(new StringReader(value));
         } catch (IOException ex) {
            LOG.error("Error converting string to properties", ex);
         }
         return properties;
      }

      @Override
      protected String valueToString(Properties value) {
         return value.toString();
      }

      @Override
      public TerminalOption<Properties> copy() {
         final TerminalOption<Properties> copy = propertiesType(name);
         copy.text = this.text;
         copy.value = new Properties(this.value);
         return copy;
      }

      @Override
      public String text() {
         final StringWriter sw = new StringWriter();
         try {
            this.value.store(sw, name);
         } catch (IOException ex) {
            LOG.error("Error converting properties to string", ex);
         }
         return sw.toString();
      }

      @Override
      public Map<String, String> textMap() {
         final Map<String, String> values = new HashMap<>(this.value.size());
         this.value.forEach((k, v) ->values.put(k.toString(), v.toString()));
         return values;
      }
   }
}
