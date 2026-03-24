/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt;

import org.rgt.options.TerminalOptionListener;
import java.net.InetAddress;
import java.net.UnknownHostException;
import java.util.ArrayList;
import java.util.Collection;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Properties;
import org.rgt.options.TerminalOption;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 *
 * @author fabio_uggeri
 */
public class TerminalServerConfiguration {

   private static final Logger LOG = LoggerFactory.getLogger(TerminalServerConfiguration.class);

   public static final String PASSTHROUGHT_AUTHENTICATION = "passthrought";

   public static final String TERMINAL_AUTHENTICATION = "terminal";

   public static final String LDAP_AUTHENTICATION = "ldap";

   private final TerminalOption<String> address = TerminalOption.stringType("server.address", "");

   private final TerminalOption<Integer> port = TerminalOption.intType("server.port", "7654");

   private final TerminalOption<Integer> adminPort = TerminalOption.intType("server.adminPort", "7656");

   private final TerminalOption<RGTLogLevel> teLogLevel = TerminalOption.logLevelType("terminal.logLevel", "WARNING");

   private final TerminalOption<String> teLogPathName = TerminalOption.stringType("terminal.logPathName", "rgt-terminal.log");

   private final TerminalOption<RGTLogLevel> serverLogLevel = TerminalOption.logLevelType("server.logLevel", "WARNING");

   private final TerminalOption<String> serverLogPathName = TerminalOption.stringType("server.logPathName", "rgt-server.log");

   private final TerminalOption<RGTLogLevel> appLogLevel = TerminalOption.logLevelType("application.logLevel", "WARNING");

   private final TerminalOption<String> appLogPathName = TerminalOption.stringType("application.logPathName",
           "logs-app/rgt-app-${pid}_${date}_${time}.log");

   private final TerminalOption<Boolean> showConsole = TerminalOption.booleanType("application.console.show", "false");

   private final TerminalOption<String> terminalAuthenticationMode = TerminalOption.stringType("terminal.authentication.mode",
           PASSTHROUGHT_AUTHENTICATION);

   private final TerminalOption<Properties> terminalAuthenticationConfig = TerminalOption.propertiesType("terminal.authentication");

   private final TerminalOption<String> adminAuthenticationMode = TerminalOption.stringType("admin.authentication.mode",
           LDAP_AUTHENTICATION);

   private final TerminalOption<Properties> adminAuthenticationConfig = TerminalOption.propertiesType("admin.authentication");

   private final Map<String, TerminalOption> options = new HashMap<>();

   private final Map<String, String> envVars = new HashMap<>();

   private String availableProcessors(int mult) {
      return Integer.toString(Runtime.getRuntime().availableProcessors() * mult);
   }

   public TerminalServerConfiguration() {
      options.put(address.name().toLowerCase(), address);
      options.put(port.name().toLowerCase(), port);
      options.put(adminPort.name().toLowerCase(), adminPort);
      options.put(teLogLevel.name().toLowerCase(), teLogLevel);
      options.put(teLogPathName.name().toLowerCase(), teLogPathName);
      options.put(serverLogLevel.name().toLowerCase(), serverLogLevel);
      options.put(serverLogPathName.name().toLowerCase(), serverLogPathName);
      options.put(appLogLevel.name().toLowerCase(), appLogLevel);
      options.put(appLogPathName.name().toLowerCase(), appLogPathName);
      options.put(showConsole.name().toLowerCase(), showConsole);
      options.put(terminalAuthenticationMode.name().toLowerCase(), terminalAuthenticationMode);
      options.put(adminAuthenticationMode.name().toLowerCase(), adminAuthenticationMode);
   }

   public void addListener(TerminalOptionListener listener) {
      options.values().forEach((t) -> t.addListener(listener));
   }

   public boolean removeListener(TerminalOptionListener listener) {
      options.values().forEach((t) -> t.removeListener(listener));
      return true;
   }

   public void clearListeners() {
      options.values().forEach((t) -> t.clearListeners());
   }

   public boolean isValidAddress() {
      try {
         return address.value().isEmpty() || InetAddress.getByName(address.value().trim()) != null;
      } catch (UnknownHostException ex) {
         LOG.warn("Invalid address informed.", ex);
      }
      return false;
   }

   public Collection<String> validate() {
      final List<String> errors = new ArrayList<>();
      if (!isValidAddress()) {
         errors.add("Invalid service address");
      }
      if (port.value() < 1 || port.value() > 65535) {
         errors.add("Port service number must be between 1 and 65535");
      }
      if (adminPort.value() < 1 || adminPort.value() > 65535) {
         errors.add("Administrarion port service number must be between 1 and 65535");
      }
      if (teLogPathName.value().isEmpty()) {
         errors.add("Invalid path and name for terminal emulator log file");
      }
      if (serverLogPathName.value().isEmpty()) {
         errors.add("Invalid path and name for server file");
      }
      if (appLogPathName.value().isEmpty()) {
         errors.add("Invalid path and name for application log file");
      }
      return errors;
   }

   public void set(TerminalOption option) {
      final TerminalOption o = options.get(option.name());
      if (o != null) {
         o.value(option.value());
      } else {
         options.put(option.name().toLowerCase(), option.copy());
      }
   }

   public void set(final TerminalServerConfiguration config) {
      config.options.forEach((k, v) -> set(v));
      envVars.putAll(config.envVars);
   }

   public Collection<TerminalOption> getOptions() {
      return options.values();
   }

   public Map<String, String> getEnvVars() {
      return envVars;
   }

   public <T> TerminalOption<T> getOption(final String name) {
      return options.get(name.toLowerCase());
   }

   public String getEnvVar(final String name) {
      final String lowerName = name.toLowerCase();
      if (lowerName.startsWith("env.")) {
         return envVars.getOrDefault(lowerName.substring(4), "");
      } else {
         return envVars.getOrDefault(lowerName, "");
      }
   }

   public <T> T get(final String name) {
      final String lowerName = name.toLowerCase();
      if (lowerName.startsWith("env.")) {
         return (T) envVars.get(lowerName.substring(4));
      } else {
         final TerminalOption<T> o = options.get(lowerName);
         if (o != null) {
            return o.value();
         }
      }
      return (T) envVars.get(lowerName);
   }

   public TerminalServerConfiguration address(final String value) {
      address.value(value);
      return this;
   }

   public TerminalOption<String> address() {
      return address;
   }

   public TerminalServerConfiguration port(final int value) {
      port.value(value);
      return this;
   }

   public TerminalOption<Integer> port() {
      return port;
   }

   public TerminalServerConfiguration adminPort(final int value) {
      adminPort.value(value);
      return this;
   }

   public TerminalOption<Integer> adminPort() {
      return adminPort;
   }

   public TerminalOption<Boolean> showConsole() {
      return showConsole;
   }

   public TerminalServerConfiguration teLogLevel(RGTLogLevel value) {
      teLogLevel.value(value);
      return this;
   }

   public TerminalOption<RGTLogLevel> teLogLevel() {
      return teLogLevel;
   }

   public TerminalServerConfiguration teLogPathName(String value) {
      teLogPathName.value(value);
      return this;
   }

   public TerminalOption<String> teLogPathName() {
      return teLogPathName;
   }

   public TerminalServerConfiguration appLogLevel(RGTLogLevel value) {
      appLogLevel.value(value);
      return this;
   }

   public TerminalOption<RGTLogLevel> appLogLevel() {
      return appLogLevel;
   }

   public TerminalServerConfiguration appLogPathName(String value) {
      appLogPathName.value(value);
      return this;
   }

   public TerminalOption<String> appLogPathName() {
      return appLogPathName;
   }

   public TerminalServerConfiguration serverLogLevel(RGTLogLevel value) {
      serverLogLevel.value(value);
      return this;
   }

   public TerminalOption<RGTLogLevel> serverLogLevel() {
      return serverLogLevel;
   }

   public TerminalServerConfiguration serverLogPathName(String value) {
      serverLogPathName.value(value);
      return this;
   }

   public TerminalOption<String> serverLogPathName() {
      return serverLogPathName;
   }

   public TerminalServerConfiguration terminalAuthenticationMode(String value) {
      terminalAuthenticationMode.value(value);
      return this;
   }

   public TerminalOption<String> terminalAuthenticationMode() {
      return terminalAuthenticationMode;
   }

   public TerminalServerConfiguration terminalAuthenticationConfig(Object key, Object value) {
      terminalAuthenticationConfig.value().put(key, value);
      return this;
   }

   public TerminalServerConfiguration terminalAuthenticationConfig(Properties value) {
      terminalAuthenticationConfig.value().clear();
      terminalAuthenticationConfig.value().putAll(value);
      return this;
   }

   public TerminalOption<Properties> terminalAuthenticationConfig() {
      return terminalAuthenticationConfig;
   }

   public TerminalServerConfiguration adminAuthenticationMode(String value) {
      adminAuthenticationMode.value(value);
      return this;
   }

   public TerminalOption<String> adminAuthenticationMode() {
      return adminAuthenticationMode;
   }

   public TerminalServerConfiguration adminAuthenticationConfig(Object key, Object value) {
      adminAuthenticationConfig.value().put(key, value);
      return this;
   }

   public TerminalServerConfiguration adminAuthenticationConfig(Properties value) {
      adminAuthenticationConfig.value().clear();
      adminAuthenticationConfig.value().putAll(value);
      return this;
   }

   public TerminalOption<Properties> adminAuthenticationConfig() {
      return adminAuthenticationConfig;
   }

   public void delete(final String name) {
      final String propName = name.toLowerCase();
      if (propName.startsWith("env.")) {
         envVars.remove(propName.substring(4));
      } else if (options.remove(propName) == null) {
         envVars.remove(propName);
      }
   }

   public void setValue(final String name, final String value) {
      final String propName = name.toLowerCase();
      if (propName.startsWith("env.")) {
         envVars.put(propName.substring(4), value);
      } else {
         for (TerminalOption option : options.values()) {
            if (option.name().equalsIgnoreCase(propName)) {
               option.text(value);
               return;
            } else if (propName.startsWith(option.name().toLowerCase()) && Properties.class.isInstance(option.value())) {
               ((Properties)option.value()).put(propName, value);
               return;
            }
         }
         options.put(propName, TerminalOption.stringType(name, value));
      }
   }

   public void setValues(Map<String, String> values) {
      values.forEach((k, v) -> setValue(k, v));
   }

   public void setValues(Properties values) {
      values.forEach((k, v) -> setValue(k.toString(), v.toString()));
   }

   public Map<String, String> stringValues() {
      final Map<String, String> values = new HashMap<>();
      options.values().forEach((t) -> values.putAll(t.textMap()));
      envVars.forEach((k, v) -> values.put("env." + k, v));
      return values;
   }
}
