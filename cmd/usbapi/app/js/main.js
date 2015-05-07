(function () {
	$(document).ready(function() {
		$('#graph-btn').click(function(e) {
			$(this).hide();
			$('#navbar-collapse-1').collapse('hide');
			$('.mypanel').hide();
			$('#mygraph').show();
		});
	});
})();


