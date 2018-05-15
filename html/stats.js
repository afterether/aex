function stats_load() {
	plot_difficulty();
}
function plot_difficulty() {

	Ajax_GET('/stats_difficulty',function(data) {
			var response=JSON.parse(data);
			var unit=response.result.Unit
			var points={};
			points.series=[]
			var difficulty={}
			difficulty.name="Difficulty"
			difficulty.data=[]
			var i
			i=0;
			for (;i<response.result.Timestamps.length;i++) {
				var date=new Date(response.result.Timestamps[i]*1000)
				var obj={}
				obj.x=date
				obj.y=response.result.Values[i]
				difficulty.data.push(obj)
			}
			points.series.push(difficulty)
			new Chartist.Line('#difficulty_chart',points, {
				lineSmooth: Chartist.Interpolation.simple({
					divisor: 2,
				    fillHoles: false
				}),
				showPoint: false,
				axisX: {
					type: Chartist.FixedScaleAxis,
					divisor: 10,
					labelInterpolationFnc: function(value) {
						return moment(value).format('MMM D');
					}
				},
				axisY: {
					labelInterpolationFnc: function(value) {
						return value + unit;
					}
				}
			});
	});

}

