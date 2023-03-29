default:
	ffmpeg -y -i verse1.mp4 -i verse2.mp4 -i verse3.mp4 -filter_complex "[0:v] [0:a] [1:v] [1:a] [2:v] [2:a] concat=n=3:v=1:a=1" -c:v libx264 -preset fast -crf 23 -c:a aac -b:a 128k green.mp4
	# ffmpeg -i "concat:verse1.mp4|verse2.mp4|verse3.mp4" -c copy green.mp4
	# ffmpeg -y -i verse2.mp4 -i verse1.mp4 -i verse3.mp4 green.mp4


