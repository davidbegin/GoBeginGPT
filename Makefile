default:
	go build -o ./bin/$(NAME) . 

remix:
	./bin/GoBeginGPT -remix -remix_id=2295844 -prompt="Office covered in Dank Weed"

skybox:
	./bin/GoBeginGPT


# These are all the folders and files we need to get started
# TODO: finish updating these
init:
	mkdir -p tmp
	mkdir -p skybox_archive
