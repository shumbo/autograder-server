package report

import (
    "fmt"
    "time"

    "gonum.org/v1/gonum/stat"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
    "github.com/eriq-augustine/autograder/usr"
)

const (
    OVERALL_NAME = "<Overall>"
)

type AssignmentScoringReport struct {
    AssignmentName string `json:"assignment-name"`
    NumberOfSubmissions int `json:"number-of-submissions"`
    LatestSubmission time.Time `json:"latest-submission"`
    Questions []*ScoringReportQuestionStats `json:"questions"`

    LatestSubmissionString string `json:"-"`
}

type ScoringReportQuestionStats struct {
    QuestionName string `json:"question-name"`

    Min float64 `json:"min"`
    Max float64 `json:"max"`
    Median float64 `json:"median"`
    Mean float64 `json:"mean"`
    StdDev float64 `json:"standard-deviation"`

    MinString string `json:"-"`
    MaxString string `json:"-"`
    MedianString string `json:"-"`
    MeanString string `json:"-"`
    StdDevString string `json:"-"`
}

const DEFAULT_VALUE float64 = -1.0;

func GetAssignmentScoringReport(assignment model.Assignment) (*AssignmentScoringReport, error) {
    questionNames, scores, lastSubmissionTime, err := fetchScores(assignment);
    if (err != nil) {
        return nil, err;
    }

    numSubmissions := 0;
    questions := make([]*ScoringReportQuestionStats, 0, len(questionNames));

    for _, questionName := range questionNames {
        min, max := util.MinMax(scores[questionName]);
        mean, stdDev := stat.MeanStdDev(scores[questionName], nil);
        median := util.Median(scores[questionName]);

        stats := &ScoringReportQuestionStats{
            QuestionName: questionName,
            Min: util.DefaultNaN(min, DEFAULT_VALUE),
            Max: util.DefaultNaN(max, DEFAULT_VALUE),
            Median: util.DefaultNaN(median, DEFAULT_VALUE),
            Mean: util.DefaultNaN(mean, DEFAULT_VALUE),
            StdDev: util.DefaultNaN(stdDev, DEFAULT_VALUE),

            MinString: fmt.Sprintf("%0.2f", min),
            MaxString: fmt.Sprintf("%0.2f", max),
            MedianString: fmt.Sprintf("%0.2f", median),
            MeanString: fmt.Sprintf("%0.2f", mean),
            StdDevString: fmt.Sprintf("%0.2f", stdDev),
        };

        questions = append(questions, stats);
        numSubmissions = len(scores[questionName]);
    }

    report := AssignmentScoringReport{
        AssignmentName: assignment.GetName(),
        NumberOfSubmissions: numSubmissions,
        LatestSubmission: lastSubmissionTime,
        LatestSubmissionString: lastSubmissionTime.Format(time.DateTime),
        Questions: questions,
    };

    return &report, nil;
}

func fetchScores(assignment model.Assignment) ([]string, map[string][]float64, time.Time, error) {
    users, err := assignment.GetCourse().GetUsers();
    if (err != nil) {
        return nil, nil, time.Time{}, fmt.Errorf("Failed to get users for course: '%w'.", err);
    }

    paths, err := assignment.GetAllRecentSubmissionResults(users);
    if (err != nil) {
        return nil, nil, time.Time{}, fmt.Errorf("Failed to get submission results: '%w'.", err);
    }

    questionNames := make([]string, 0);
    scores := make(map[string][]float64);
    lastSubmissionTime := time.Time{};

    for email, path := range paths {
        if (users[email].Role != usr.Student) {
            continue;
        }

        if (path == "") {
            continue;
        }

        result := artifact.GradedAssignment{};
        err = util.JSONFromFile(path, &result);
        if (err != nil) {
            return nil, nil, time.Time{}, fmt.Errorf("Failed to deserialize submission result '%s': '%w'.", path, err);
        }

        if (result.GradingStartTime.After(lastSubmissionTime)) {
            lastSubmissionTime = result.GradingStartTime;
        }

        if (len(questionNames) == 0) {
            for _, question := range result.Questions {
                questionNames = append(questionNames, question.Name);
                scores[question.Name] = make([]float64, 0);
            }

            questionNames = append(questionNames, OVERALL_NAME);
        }

        total := 0.0
        max_points := 0.0

        for _, question := range result.Questions {
            var score float64 = 0.0;
            if (!util.IsZero(question.MaxPoints)) {
                score = question.Score / question.MaxPoints;
            }

            scores[question.Name] = append(scores[question.Name], score);

            total += question.Score;
            max_points += question.MaxPoints;
        }

        total_score := 0.0;
        if (!util.IsZero(max_points)) {
            total_score = total / max_points;
        }


        scores[OVERALL_NAME] = append(scores[OVERALL_NAME], total_score);
    }

    return questionNames, scores, lastSubmissionTime, nil;
}
