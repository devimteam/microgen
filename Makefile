install: ; go install ./cmd/microgen

examples_update:
	cd ./examples   ;\
	echo addsvc     ;\
	cd ./addsvc     ;\
	rm -rf ./cmd ./service ./transport   ;\
	make gen        ;\
	cd ..           ;\
	echo usersvc    ;\
	cd ./usersvc    ;\
    rm -rf ./cmd ./service ./transport   ;\
    make gen        ;\
    cd ..           ;\
	cd ..           ;\
