from datetime import datetime
from pathlib import Path
import pandas as pd


def main():
    all_results = []
    for file in (Path.cwd() / "load_tests/results/csv").glob("*.csv"):
        results = process_file(file)
        results["filename"] = file.name
        all_results.append(results)
    df = pd.DataFrame(all_results)
    df.to_csv("results.csv", index=False)
    df.T.to_csv("resultsT.csv")


def format_bytes(size):
    power = 2**10
    n = 0
    power_labels = {0: "", 1: "К", 2: "М", 3: "Г"}
    while size > power:
        size /= power
        n += 1
    return size, power_labels[n] + "Б"


def process_file(file: Path):
    print(f"\nProcessing {file.name}...")
    df = pd.read_csv(file)

    results = {}

    req_df = df[df["metric_name"] == "http_req_duration"]
    results["Среднее время запроса"] = round(req_df["metric_value"].mean(), 2)
    results["Максимальное время запроса"] = round(req_df["metric_value"].max(), 2)

    checks_df = df[df["metric_name"] == "checks"].groupby("check")
    metric_checks = checks_df.groups.keys()

    for check in metric_checks:
        checks_total = df[df["check"] == check]
        check_percent = checks_total["metric_value"].sum() / checks_total["metric_value"].count()
        results[f"Проверка ({check})"] = check_percent

    total_reqs = req_df["metric_value"].count()
    threshold1 = req_df[req_df["metric_value"] > 1000]["metric_value"].count() / total_reqs
    results["Запросы дольше 1000ms < 5%"] = True if threshold1 < 0.05 else False

    req_failed_df = df[df["metric_name"] == "http_req_failed"]
    threshold2 = req_failed_df[req_failed_df["metric_value"] == 1.0]["metric_value"].count() / total_reqs
    results["Процент проваленных запросов < 5%"] = True if threshold2 < 0.05 else False

    data_sent_df = df[df["metric_name"] == "data_sent"]
    data_sent, label_sent = format_bytes(data_sent_df["metric_value"].sum())

    data_received_df = df[df["metric_name"] == "data_received"]
    data_received, label_received = format_bytes(data_received_df["metric_value"].sum())
    results["Отправлено данных"] = f"{round(data_sent, 1)} {label_sent}"
    results["Получено данных"] = f"{round(data_received, 1)} {label_received}"

    time_start = datetime.fromtimestamp(req_df["timestamp"].min())
    time_end = datetime.fromtimestamp(req_df["timestamp"].max())
    time_total = (time_end - time_start).seconds
    results["RPS"] = round(total_reqs / time_total)

    for name, result in results.items():
        print(f"{name}: {result}")

    return results


if __name__ == "__main__":
    main()
