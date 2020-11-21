
tf:
	go test ./

tb:
	go test ./ -bench=. -benchmem

t: tf tb
	
