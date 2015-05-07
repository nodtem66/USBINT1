(function () {
  //Start CLOSURE
  var app = angular.module('mainApp');
  /* ---------------- Function Library -------------------------------------------------- */
  app.factory('func', function funcFactory($filter) {
    var self = this;

    this.showOption = function(text, option) {
      var text_name = arguments[2] || "text";
      var value_name = arguments[3] || "value";
      var temp = {};

      temp[value_name] = text;
      var selected = $filter('filter')(option, temp);
      return (selected.length) ? selected[0][text_name] : 'Not Set';
    };

    /*
     * 
     * @param {String} text
     * @param {String} option
     * @param {String} text_name
     * @param {String} value_name
     * @returns {Array|undefined}
     *       {id: 'object_id_class', option: $scope.classModel, text: 'objectClass', value: 'id'}
     */
    this.getOption = function(text, option) {
      var text_name = arguments[2] || "text";
      var value_name = arguments[3] || "value";
      var selected = [];

      if (text && text.toLowerCase) {
        for (var i = 0, len = option.length; i < len; i++) {
          if (option[i][text_name].toLowerCase().search(text.toLowerCase()) != -1) {
            selected.push(option[i][value_name]);
          }
        }
      }
      if (selected.length) {
        return selected;
      }
      return;
    };

    this.checkNotNull = function(data) {
      if (data) {
        return true;
      }
      return "Invalid value: Empty";
    };
    this.formatSpringJson = function(data) {
      if (data) {
        delete data['$$hashKey'];
        for (var key in data) {
          if (key) {
            var matched = key.match(/^is([A-Z])/);
            if (matched && matched.length && matched.length > 1) {
              var new_key = key.replace(matched[0], matched[1].toLowerCase());
              data[new_key] = data[key];
              delete data[key];
            }
          }
        }
      }
      return data;
    };
    this.springJsonEncoder = function(data, getHeaders) {
      var headers = getHeaders();
      //headers[ "Content-type" ] = "application/json; charset=utf-8";
      return JSON.stringify(self.formatSpringJson(data));
    };

    // A request-transformation method that is used to prepare the outgoing
    // request as a FORM post instead of a JSON packet.
    // From: http://www.bennadel.com/blog/2615-posting-form-data-with-http-in-angularjs.htm
    this.formEncoder = function(data, getHeaders) {
      var headers = getHeaders();
      //headers[ "Content-type" ] = "application/x-www-form-urlencoded; charset=utf-8";
      delete data['$$hashKey'];
      return serializeData(data);
    };
    // ---
    // PRVIATE METHODS.
    // ---


    // I serialize the given Object into a key-value pair string. This
    // method expects an object and will default to the toString() method.
    // --
    // NOTE: This is an atered version of the jQuery.param() method which
    // will serialize a data collection for Form posting.
    // --
    // https://github.com/jquery/jquery/blob/master/src/serialize.js#L45
    function serializeData(data) {

      // If this is not an object, defer to native stringification.
      if (!angular.isObject(data)) {

        return((data == null) ? "" : data.toString());

      }

      var buffer = [];

      // Serialize each key in the object.
      for (var name in data) {

        if (!data.hasOwnProperty(name)) {

          continue;

        }

        var value = data[ name ];

        buffer.push(
                encodeURIComponent(name) +
                "=" +
                encodeURIComponent((value == null) ? "" : value)
                );

      }

      // Serialize the buffer and clean it up for transportation.
      var source = buffer
              .join("&")
              .replace(/%20/g, "+")
              ;

      return(source);
    }



    return this;
  });

  //END CLOSURE
})();