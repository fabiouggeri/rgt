/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol.admin.server;

import java.util.HashMap;
import java.util.Map;
import org.rgt.protocol.Response;
import org.rgt.admin.AdminResponseCode;

/**
 *
 * @author fabio_uggeri
 */
public class GetConfigResponse extends Response {

   final private HashMap<String, String> config = new HashMap<>();

   public GetConfigResponse(AdminResponseCode responseCode, String message) {
      super(responseCode, message);
   }

   public GetConfigResponse(AdminResponseCode responseCode) {
      super(responseCode, null);
   }

   public GetConfigResponse() {
      this(AdminResponseCode.SUCCESS, null);
   }

   public GetConfigResponse(final Map<String, String> config) {
      this();
      this.config.putAll(config);
   }

   public Map<String, String> config() {
      return config;
   }

   public GetConfigResponse config(Map<String, String> otherValues) {
      this.config.putAll(otherValues);
      return this;
   }

   public GetConfigResponse put(final String key, final String value) {
      config.put(key, value);
      return this;
   }

}
