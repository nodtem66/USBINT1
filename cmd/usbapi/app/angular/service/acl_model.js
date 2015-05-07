(function() {
  //Start CLOSURE
  var app = angular.module('mainApp');
  /* ------ Model -------------------------------------------------------
   *
   */
  app.factory('acl_model', function AclModelFactory(function_queue) {
    this.sid_data = [];
    this.class_data = [];
    this.objectid_data = [];
    this.entry_data = [];
    
    /*/
     this.sid_data = [
     {id: 1, sid: "test 1", isPrincipal: false},
     {id: 2, sid: "test 2", isPrincipal: false},
     {id: 3, sid: "test 3", isPrincipal: false},
     {id: 4, sid: "test 4", isPrincipal: false},
     {id: 5, sid: "test 5", isPrincipal: true},
     {id: 6, sid: "test 6", isPrincipal: false}
     ];
     this.class_data = [
     {id: 1, objectClass: "th.ac.mwit.test.1"},
     {id: 2, objectClass: "th.ac.mwit.product.1"},
     {id: 3, objectClass: "th.ac.mwit.jirawat.example"},
     {id: 4, objectClass: "th.ac.mwit.acl.test"},
     {id: 5, objectClass: "th.ac.mwit.test.project"},
     {id: 6, objectClass: "th.ac.mwit.spring.lib"}
     ];
     this.objectid_data = [
     {id: 1, object_id_class: 2, object_id_identity: 12, parent_object: undefined, owner_sid: 1, isEntriesInherriting: false},
     {id: 2, object_id_class: 2, object_id_identity: 12, parent_object: undefined, owner_sid: 1, isEntriesInherriting: false},
     {id: 3, object_id_class: 2, object_id_identity: 12, parent_object: undefined, owner_sid: 1, isEntriesInherriting: false},
     {id: 4, object_id_class: 2, object_id_identity: 12, parent_object: undefined, owner_sid: 1, isEntriesInherriting: false},
     {id: 5, object_id_class: 2, object_id_identity: 12, parent_object: undefined, owner_sid: 1, isEntriesInherriting: false},
     {id: 6, object_id_class: 2, object_id_identity: 12, parent_object: undefined, owner_sid: 1, isEntriesInherriting: false},
     {id: 7, object_id_class: 2, object_id_identity: 12, parent_object: undefined, owner_sid: 1, isEntriesInherriting: false}
     ];
     this.entry_data = [
     {id: 1, mask: 1, aclObjectentity: 1, aceOrder: 1, sid: 1, isGranting: true},
     {id: 2, mask: 1, aclObjectentity: 2, aceOrder: 1, sid: 1, isGranting: true},
     {id: 3, mask: 1, aclObjectentity: 3, aceOrder: 1, sid: 1, isGranting: false},
     {id: 4, mask: 1, aclObjectentity: 4, aceOrder: 1, sid: 1, isGranting: false},
     {id: 5, mask: 1, aclObjectentity: 5, aceOrder: 1, sid: 1, isGranting: true},
     {id: 6, mask: 1, aclObjectentity: 6, aceOrder: 1, sid: 1, isGranting: false},
     {id: 7, mask: 1, aclObjectentity: 7, aceOrder: 1, sid: 1, isGranting: true}
     ];
     //*/

    this.count = {sid: 0, class: 0, entry: 0, objectid: 0};
    this.current_rownum = {sid: 0, class: 0, entry: 0, objectid: 0};

    //----- Helper Function to Fetch Data Async --------
    var self = this;
    var getAll = function(column) {
      return function($http) {
        //console.log(self['current_rownum'][column]);
        var start = self['current_rownum'][column] + 1;
        var count = self['count'][column];
        if (count >= start) {
          $http.get(column, {params: {'start': start, 'limit': 50}}).then(function(result) {
            if (result.data) {
              [].push.apply(self[column + '_data'], result.data);
              self['current_rownum'][column] += result.data.length;
              //console.log(column+": ", result.data);
            }
            function_queue.denyResult();
          }, function() {
            //console.log("error");
            function_queue.denyResult();
          });
        } else {
          //console.log(column+ " finish");
          function_queue.acceptResult();
        }
      };
    };

    var getCount = function(column) {
      return function($http) {
        $http.get(column + '/count').then(function(result) {
          if (result.data && result.data.count) {
            self['count'][column] = result.data.count;
            function_queue.push.call(self, getAll(column));
            function_queue.acceptResult();
          } else {
            function_queue.denyResult();
          }
        }, function() {
          function_queue.denyResult();
        });
      };
    };

    function_queue.push(getCount('sid'));
    function_queue.push(getCount('class'));
    function_queue.push(getCount('entry'));
    function_queue.push(getCount('objectid'));

    return this;
  });
  //END CLOSURE
})();