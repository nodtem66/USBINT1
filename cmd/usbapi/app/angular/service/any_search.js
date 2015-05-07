(function () {
  //Start CLOSURE
  var app = angular.module('mainApp');
  //----- Filters -------------------------------------------/
  app.filter('any_search', ['$filter', 'func', function($filter, func) {
      return function(items, text) {
        var names = arguments[2] || [];
        var item1 = $filter('filter')(items, this.search);

        for (var k = 0, len = names.length; k < len; k++) {
          var custom_search = {};
          var selected_list = func.getOption(text, names[k]['option'], names[k]['text'], names[k]['value']);
          for (var index in selected_list) {
            if (!index) {
              continue;
            }
            custom_search[names[k]['id']] = selected_list[index];
            var item2 = $filter('filter')(items, custom_search, true);
            if (item1 && item2) {
              for (var i = 0, j = 0, nlen = item2.length, mlen = item1.length; i < nlen; i++) {
                for (j = 0; j < mlen; j++) {
                  if (item1[j].id == item2[i].id) {
                    break;
                  }
                }
                if (custom_search[names[k]['id']] != undefined && j == mlen) {
                  item1.push(item2[i]);
                }
              } // End for loop item1 and item2

            } // End if item1 && item2 
          }  // End for loop selected_list
        } //End for loop names
        return item1;
      };
    }]);

  //END CLOSURE
})();

