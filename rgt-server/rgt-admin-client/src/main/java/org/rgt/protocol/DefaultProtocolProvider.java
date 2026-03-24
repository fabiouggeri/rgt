/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol;

import java.lang.reflect.Constructor;
import java.lang.reflect.InvocationTargetException;
import java.util.HashMap;
import java.util.Map.Entry;
import java.util.TreeMap;
import org.reflections.Reflections;
import org.rgt.Operation;
import org.rgt.admin.AdminOperation;
import org.rgt.protocol.admin.DefaultAdminProtocol;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 *
 * @author fabio_uggeri
 */
public class DefaultProtocolProvider implements ProtocolProvider {

   private static final Logger LOG = LoggerFactory.getLogger(DefaultProtocolProvider.class);

   private final HashMap<Operation, TreeMap<Short, Protocol<?, ?>>> operations = new HashMap<>();

   public DefaultProtocolProvider() {
      registerOperations();
   }

   private void registerOperations() {
      final Reflections reflections = new Reflections("org.rgt");
      LOG.info("Scanning protocols...");
      reflections.getTypesAnnotatedWith(AdminProtocolRegister.class).forEach(c -> registerAdminOperation(c));
      registerAdminDefaultProtocol();
   }

   private void registerAdminOperation(Class<?> protocolClass) {
      final AdminProtocolRegister annotation = protocolClass.getAnnotation(AdminProtocolRegister.class);
      if (annotation != null) {
         for (AdminOperation op : annotation.operation()) {
            registerProtocolVersion(op, annotation.fromVersion(), protocolClass);
         }
      } else {
         LOG.warn("Register annotation not found in class {}", protocolClass.getName());
      }
   }

   private void registerProtocolVersion(final Operation operation, final short protocolVersion, Class<?> protocolClass) {
      final TreeMap<Short, Protocol<?, ?>> map = operationsVersionsMap(operation);
      final Entry<Short, Protocol<?, ?>> entry = map.floorEntry(protocolVersion);
      if (entry == null || entry.getKey() != protocolVersion) {
         try {
            map.put(protocolVersion, createProtocolInstance(operation, protocolClass));
            LOG.info("Protocol version {} registered for operation {}", protocolVersion, operation);
         } catch (InstantiationException
                 | IllegalAccessException
                 | IllegalArgumentException
                 | InvocationTargetException
                 | SecurityException
                 | NoSuchMethodException ex) {
            LOG.error("Error creating protocol version " + protocolVersion + " for operation " + operation, ex);
         }
      } else {
         LOG.error("Protocol version {} not found for operation {}", protocolVersion, operation);
      }
   }

   private Protocol<?, ?> createProtocolInstance(Operation operation, Class<?> protocolClass) throws NoSuchMethodException
           , InstantiationException, IllegalAccessException, IllegalArgumentException, InvocationTargetException {
      Constructor<?> constructor;
      Protocol<?, ?> protocol;
      try {
         constructor = protocolClass.getDeclaredConstructor(operation.getClass());
         protocol = (Protocol<?, ?>) constructor.newInstance(operation);
      } catch (NoSuchMethodException ex) {
         constructor = protocolClass.getDeclaredConstructor();
         protocol = (Protocol<?, ?>) constructor.newInstance();
      }
      return protocol;
   }

   private TreeMap<Short, Protocol<?, ?>> operationsVersionsMap(final Operation operation) {
      TreeMap<Short, Protocol<?, ?>> map = operations.get(operation);
      if (map == null) {
         map = new TreeMap<>();
         operations.put(operation, map);
      }
      return map;
   }

   @Override
   public <P extends Protocol<? extends Request, ? extends Response>> P getProtocol(Operation operation, short version)
           throws ProtocolErrorException {
      final TreeMap<Short, Protocol<?, ?>> protocolVersions = operations.get(operation);
      final Entry<Short, Protocol<?, ?>> entry;
      if (protocolVersions == null) {
         throw new ProtocolErrorException("Client and server incompatibility. Unsupported operation: " + operation);
      }
      if (version < 1 || version > operation.getMaxVersion()) {
         throw new ProtocolErrorException("Client and server incompatibility. Protocol version not supported: " + version);
      }
      entry = protocolVersions.floorEntry(version);
      if (entry == null) {
         throw new ProtocolErrorException("Client and server incompatibility. Protocol version not supported: " + version);
      }
      return (P) entry.getValue();
   }

   private void registerAdminDefaultProtocol() {
      for (AdminOperation op : AdminOperation.values()) {
         if (! operations.containsKey(op)) {
            final TreeMap<Short, Protocol<?, ?>> map = new TreeMap<>();
            operations.put(op, map);
            map.put((short)1, new DefaultAdminProtocol(op));
            LOG.info("Default protocol version {} registered for operation {}", 1, op);
         }
      }
   }
}
