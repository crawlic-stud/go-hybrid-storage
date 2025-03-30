
test:
	bash load_tests/run_with_backend.bash $$back $$test

test-all:
	for back in mongo sqlite postgres fs ; do \
		for test in get_all get_random_file upload_small_chunk upload_large_chunk ; do \
			bash load_tests/run_with_backend.bash $$back $$test ; \
			echo "_______________________________________________________________" ; \
			echo "sleeping a minute between tests..." ; \
			sleep 60 ; \
		done ; \
	done

	echo "All tests finished."
	echo "Plotting results..."
	python load_tests/plot_results.py 
	echo "Script finished."

test-back:
	for test in get_all get_random_file upload_small_chunk upload_large_chunk ; do \
		bash load_tests/run_with_backend.bash $$back $$test ; \
		echo "_______________________________________________________________" ; \
		echo "sleeping a minute between tests..." ; \
		sleep 60 ; \
	done 

	echo "All tests finished."
	echo "Plotting results..."
	python load_tests/plot_results.py $$back
	echo "Script finished."

