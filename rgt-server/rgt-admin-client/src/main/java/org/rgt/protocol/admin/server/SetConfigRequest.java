/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.server;

import java.util.HashMap;
import java.util.Map;
import org.rgt.protocol.Request;
import org.rgt.admin.AdminOperation;

/**
 *
 * @author fabio_uggeri
 */
public class SetConfigRequest extends Request {

   private final HashMap<String, String> config = new HashMap<>();

   private boolean removeMissingOptions = false;

   public SetConfigRequest() {
      super(AdminOperation.SET_CONFIG);
   }

   public SetConfigRequest(final Map<String, String> config, boolean removeMissing) {
      super(AdminOperation.SET_CONFIG);
      this.config.putAll(config);
      this.removeMissingOptions = removeMissing;
   }

   public Map<String, String> config() {
      return config;
   }

   public SetConfigRequest config(final Map<String, String> otherConfig) {
      config.putAll(otherConfig);
      return this;
   }

   public SetConfigRequest put(final String key, final String value) {
      config.put(key, value);
      return this;
   }

   public SetConfigRequest removeMissingOptions(final boolean remove) {
      this.removeMissingOptions = remove;
      return this;
   }

   public boolean removeMissingOptions() {
      return removeMissingOptions;
   }
}
