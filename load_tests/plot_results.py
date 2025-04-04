"""Taken from: https://github.com/topnax/k6-results-visualization"""

import csv
from functools import lru_cache
from pathlib import Path
from dateutil import parser
import matplotlib
import matplotlib.pyplot as plt
import datetime
import json
import sys


def round_seconds(date_time_object):
    new_date_time = date_time_object

    if new_date_time.microsecond >= 500000:
        new_date_time = new_date_time + datetime.timedelta(seconds=1)

    return new_date_time.replace(microsecond=0)


def round_minutes(date_time_object):
    new_date_time = date_time_object

    if new_date_time.second >= 30:
        new_date_time = new_date_time + datetime.timedelta(minutes=1)

    return new_date_time.replace(second=0)


@lru_cache
def load_data(file_name):
    data = {}
    with open(file_name, "r") as f:
        reader = csv.DictReader(f)
        for line_data in reader:
            metric = line_data["metric_name"]
            metric_data = data.get(metric, [])
            dt = datetime.datetime.fromtimestamp(float(line_data["timestamp"]))
            metric_data.append((float(line_data["metric_value"]), dt))
            data[metric] = metric_data
    return data


def get_metric_data(data, watched_metric):
    metric_data = data[watched_metric]

    return ([time for (value, time) in metric_data], [value for (value, time) in metric_data])


def get_avg_and_max_from_data(metric_data_raw):
    # prepare dicts for (sum, count) and max
    metric_data_sum_count = {}
    metric_data_max = {}

    for time, value in zip(metric_data_raw[0], metric_data_raw[1]):
        # round the time to seconds
        # time = round_minutes(round_seconds(time))
        time = round_seconds(time)

        # find the current entry
        entry = metric_data_sum_count.get(time, (0, 0))
        # and update its sum and count
        metric_data_sum_count[time] = (entry[0] + value, entry[1] + 1)

        # find the current maximum value
        max_wait = metric_data_max.get(time, 0)
        # and update it
        metric_data_max[time] = max(max_wait, value)

    # compute data avg
    metric_data_avg = [val / count for (_, (val, count)) in metric_data_sum_count.items()]

    return (metric_data_avg, metric_data_max)


def display_chart(file_name: Path, metric_data_avg, metric_data_max, vus_data, watched_metric):
    # the chart will have two Y axes
    fig, ax1 = plt.subplots()

    # x axis displays time
    ax1.set_xlabel("Time")

    # first Y axis displays duration in milliseconds
    ax1.set_ylabel(watched_metric + " (ms)")

    # plot avg metric data values
    plot_1 = ax1.plot(metric_data_max.keys(), metric_data_avg, color="blue", label=f"avg {watched_metric}")

    # plot max metric data values
    plot_1 = ax1.plot(
        metric_data_max.keys(), metric_data_max.values(), color="red", linestyle="dashed", label=f"max {watched_metric}"
    )

    # display legend of the first Y axis
    ax1.legend(loc=1)

    # make a second Y axis that will display the number of active VUs
    ax2 = ax1.twinx()
    ax2.set_ylabel("VUs")

    # plot the VU count data
    plot_2 = ax2.plot(vus_data[0], vus_data[1], color="green", label="VU count")

    # display legend of the second Y axis
    ax2.legend(loc=2)

    # do autoformat the X axis
    plt.gcf().autofmt_xdate()

    # show the chart
    plt.title("k6 load test results chart")

    plots_path = Path.cwd() / "load_tests" / "results" / "plots" / watched_metric
    plots_path.mkdir(parents=True, exist_ok=True)
    plt.savefig(plots_path / file_name.name.replace(".csv", ".png"), dpi=300)


def process_file(file_name, watched_metric):
    # load data from the file
    print(f"Processing {file_name} for {watched_metric}...")
    data = load_data(file_name)
    print(f"File processed...")

    print("Parsing data...")
    # get raw metric data
    metric_data_raw = get_metric_data(data, watched_metric)

    # get avg and max values from data
    (metric_data_avg, metric_data_max) = get_avg_and_max_from_data(metric_data_raw)

    # get VU count data
    vus_data = get_metric_data(data, "vus")

    print("Displaying chart...")

    # display a chart
    display_chart(file_name, metric_data_avg, metric_data_max, vus_data, watched_metric)

    # display_hist(metric_data_raw[1])


if __name__ == "__main__":
    back_postfix = sys.argv[1] if len(sys.argv) > 1 else ""
    path = Path.cwd() / "load_tests" / "results" / "csv"
    files = path.glob(f"*{back_postfix}.csv")
    for file in files:
        for metric in ("http_req_waiting", "http_req_duration", "http_req_failed"):
            process_file(file, metric)
