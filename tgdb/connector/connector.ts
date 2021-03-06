/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
import {Injectable} from "@angular/core";
import {WiContrib, WiServiceHandlerContribution, AUTHENTICATION_TYPE} from "wi-studio/app/contrib/wi-contrib";
import {IConnectorContribution, IFieldDefinition, IActionResult, ActionResult} from "wi-studio/common/models/contrib";
import {Observable} from "rxjs/Observable";
import {IValidationResult, ValidationResult, ValidationError} from "wi-studio/common/models/validation";

@WiContrib({})
@Injectable()
export class TibcoGraphDatabaseConnectorContribution extends WiServiceHandlerContribution {
    constructor() {
        super();
    }

    value = (fieldName: string, context: IConnectorContribution): Observable<any> | any => {
        return null;
    }

    validate = (name: string, context: IConnectorContribution): Observable<IValidationResult> | IValidationResult => {
		console.log('------------- validate --------------');
		console.log(context);
		console.log('%%%%%%%%%% name = ' + name);
		
		if(name === "Connect") {
			let url: string;
			let username: string; 
			let password: string; 

			for (let configuration of context.settings) {
				if( configuration.name === "url") {
					url = configuration.value	
				} else if( configuration.name === "username") {
					username = configuration.value
				} else if( configuration.name === "password") {
					password = configuration.value
				}
			}

			if( url && password ) {
				return ValidationResult.newValidationResult().setReadOnly(false);
			} else {
				return ValidationResult.newValidationResult().setReadOnly(true);

			}
		}
 		return null;
    }

    action = (actionName: string, context: IConnectorContribution): Observable<IActionResult> | IActionResult => {
		if (actionName == "Connect") {
			console.log('------------- action click connect --------------');
            return Observable.create(observer => {
			console.log('------------- action callback --------------');
                /**
                 * Set the action result to save the configuration data
                 */
                let actionResult = {
                    context: context,
                    authType: AUTHENTICATION_TYPE.BASIC,
                    authData: {}
                }

                /**
                 * Call the observer and tell it the validation was sucessful and the data should be saved
                 */
                observer.next(ActionResult.newActionResult().setSuccess(true).setResult(actionResult));
            });
        }
		
		return null;
    }
}