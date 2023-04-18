package kanban

import (
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/datapi/src/test/factory"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/signaux-faibles/libwekan"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func initReferentiel() {
	// set up referentiel
	core.Departements = map[core.CodeDepartement]string{
		"75": "Paris",
		"77": "Seine et Marne",
		"59": "Nord",
		"62": "Somme",
	}
	core.Regions = map[core.Region][]core.CodeDepartement{
		"Ile de France":   {"75", "77"},
		"Hauts-de-France": {"59", "62"},
		"France entière":  {"59", "62", "75", "77"},
	}
	log.Printf(
		"référentiel initialisé (Régions : %s), (Départements : %s)\n",
		utils.GetKeys(core.Regions),
		utils.GetKeys(core.Departements),
	)
}

func Test_kanbanConfigForUser_withActiveMembers(t *testing.T) { // GIVEN
	ass := assert.New(t)
	userOne := factory.OneWekanUser()
	userTwo := factory.OneWekanUser()
	userThree := factory.OneWekanUser()
	configBoardA := factory.OneConfigBoardWithMembers(userOne, userTwo)
	configBoardB := factory.OneConfigBoardWithMembers(userOne, userThree)

	// set up wekanConfig for test
	core.WekanConfig = factory.LibwekanConfigWith(
		[]libwekan.ConfigBoard{configBoardA, configBoardB},
		[]libwekan.User{userOne, userTwo, userThree},
	)

	// WHEN
	configForUserOne := kanbanConfigForUser(userOne.Username)
	configForUserTwo := kanbanConfigForUser(userTwo.Username)
	configForUserThree := kanbanConfigForUser(userThree.Username)

	// THEN
	ass.Equal(userOne.ID, configForUserOne.UserID)
	configShouldExactlyContains(ass, configForUserOne, configBoardA.Board, configBoardB.Board)

	ass.Equal(userTwo.ID, configForUserTwo.UserID)
	configShouldExactlyContains(ass, configForUserTwo, configBoardA.Board)

	ass.Equal(userThree.ID, configForUserThree.UserID)
	configShouldExactlyContains(ass, configForUserThree, configBoardB.Board)

}

func Test_kanbanConfigForUser_assertThatAllUsersArePresents(t *testing.T) { // GIVEN
	ass := assert.New(t)
	userOne := factory.OneWekanUser()
	userLambda1 := factory.OneWekanUser()
	userLambda2 := factory.OneWekanUser()
	userLambda3 := factory.OneWekanUser()

	configBoardA := factory.OneConfigBoardWithMembers(userOne, userLambda1)
	configBoardB := factory.OneConfigBoardWithMembers(userOne, userLambda2, userLambda3)
	factory.DeactiveMembers(configBoardB, userLambda3)

	// set up wekanConfig for test
	core.WekanConfig = factory.LibwekanConfigWith(
		[]libwekan.ConfigBoard{configBoardA, configBoardB},
		[]libwekan.User{userOne, userLambda1, userLambda2, userLambda3},
	)

	// WHEN
	configForUserOne := kanbanConfigForUser(userOne.Username)

	// THEN
	ass.Equal(userOne.ID, configForUserOne.UserID)
	ass.ElementsMatch(
		utils.GetKeys(configForUserOne.Users),
		[]libwekan.UserID{userLambda1.ID, userLambda2.ID, userLambda3.ID},
	)
	ass.ElementsMatch(
		utils.GetValues(configForUserOne.Users),
		[]libwekan.Username{userLambda1.Username, userLambda2.Username, userLambda3.Username},
	)

}

func Test_kanbanConfigForUser_withInactiveMembers(t *testing.T) { // GIVEN
	ass := assert.New(t)
	userOne := factory.OneWekanUser()
	configBoardA := factory.OneConfigBoardWithMembers(userOne)
	factory.DeactiveMembers(configBoardA, userOne)

	// set up wekanConfig for test
	core.WekanConfig = factory.LibwekanConfigWith(
		[]libwekan.ConfigBoard{configBoardA},
		[]libwekan.User{userOne},
	)

	// WHEN
	configForUserOne := kanbanConfigForUser(userOne.Username)

	// THEN
	ass.Equal(userOne.ID, configForUserOne.UserID)
	ass.Empty(configForUserOne.Boards)
}

func Test_kanbanConfigForUser_hasOnlyDepartementsFromSwimlanes(t *testing.T) {
	ass := assert.New(t)

	initReferentiel()
	user := factory.OneWekanUser()

	tests := []struct {
		name string
		args []string
		want []core.CodeDepartement
	}{
		{
			"test departements",
			[]string{"59 (Nord)", "75 (Paris)"},
			[]core.CodeDepartement{core.CodeDepartement("59"), core.CodeDepartement("75")},
		},
		{
			"test région",
			[]string{"Ile de France"},
			[]core.CodeDepartement{core.CodeDepartement("75"), core.CodeDepartement("77")},
		},
		{
			"test France entière",
			[]string{"France entière"},
			[]core.CodeDepartement{core.CodeDepartement("75"), core.CodeDepartement("77"), core.CodeDepartement("62"), core.CodeDepartement("59")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configGenerale := factory.OneConfigBoardWithMembers(user)
			factory.AddSwimlanesWithDepartments(&configGenerale, tt.args...)

			// set up wekanConfig for t
			core.WekanConfig = factory.LibwekanConfigWith(
				[]libwekan.ConfigBoard{configGenerale},
				[]libwekan.User{user},
			)
			// WHEN
			configForUserOne := kanbanConfigForUser(user.Username)
			actual := utils.GetKeys(configForUserOne.Departements)
			expected := make([]core.CodeDepartement, len(tt.want))
			for i, dpt := range tt.want {
				expected[i] = dpt
			}
			// THEN
			ass.ElementsMatch(expected, actual)
		})
	}
}

func Test_parseSwimlaneTitle(t *testing.T) {
	type args struct {
		swimlane core.KanbanSwimlane
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test departement avec les parenthèses", args{core.KanbanSwimlane{Title: "93 (Seine saint-Denis)"}}, "93"},
		{"test departement sans les parenthèses", args{core.KanbanSwimlane{Title: "93"}}, "93"},
		{"test departement avec un espace final", args{core.KanbanSwimlane{Title: "93 "}}, "93"},
		{"test région avec des espaces", args{core.KanbanSwimlane{Title: "Île de France"}}, "Île de France"},
		{"test région - trim", args{core.KanbanSwimlane{Title: " Île de France "}}, "Île de France"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, parseSwimlaneTitle(tt.args.swimlane), "parseSwimlaneTitle(%v)", tt.args.swimlane)
		})
	}
}

func configShouldExactlyContains(ass *assert.Assertions, config core.KanbanConfig, boards ...libwekan.Board) {
	boardIDsFromConfig := utils.GetKeys(config.Boards)

	expectedBoardIDs := utils.Convert(boards, func(board libwekan.Board) libwekan.BoardID {
		return board.ID
	})
	ass.ElementsMatch(boardIDsFromConfig, expectedBoardIDs)
}
