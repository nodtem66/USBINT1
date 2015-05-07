(function (){
	var app = angular.module('WebMonitor');

	app.directive('copyright', function factory() {
		return {
			restrict: 'AE',
			replace: true,
			templateUrl: 'js/app/directive/copyright.html',
			controller: function($scope){
				$scope.update_time = '2015-3-11';
				$scope.version = '0.1.0'
			}
		}
	});

	app.directive('scroll', function factory($window) {
		var dom = function(scope, element, attr) {
			var offset = attr['scroll'] || 61;
			angular.element($window).bind("scroll", function() {
				if (this.pageYOffset >= offset) {
					scope.isScrollDown = true;
				} else {
					scope.isScrollDown = false;
				}
				scope.$apply();
			});
		};

		return {
			restrict: 'AE',
			replace: true,
			link: dom
		};
	});

}());