/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package org.rgt.protocol;

import static java.lang.annotation.ElementType.TYPE;
import java.lang.annotation.Retention;
import java.lang.annotation.Target;
import org.rgt.admin.AdminOperation;

/**
 *
 * @author fabio_uggeri
 */
@Retention(java.lang.annotation.RetentionPolicy.RUNTIME)
@Target(TYPE)
public @interface AdminProtocolRegister {

   AdminOperation[] operation();

   short fromVersion() default 1;

}
