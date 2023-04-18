package core

import (
	"github.com/signaux-faibles/datapi/src/test/factory"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/signaux-faibles/libwekan"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func initReferentiel() {
	// set up referentiel
	departements = map[CodeDepartement]string{
		"75": "Paris",
		"77": "Seine et Marne",
		"59": "Nord",
		"62": "Somme",
	}
	regions = map[Region][]CodeDepartement{
		"Ile de France":   {"75", "77"},
		"Hauts de France": {"59", "62"},
		"France entière":  {"59", "62", "75", "77"},
	}
	log.Printf(
		"référentiel initialisé (Régions : %s), (Départements : %s)\n",
		utils.GetKeys(regions),
		utils.GetKeys(departements),
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
	wekanConfig = factory.LibwekanConfigWith(
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

func Test_kanbanConfigForUser_withInactiveMembers(t *testing.T) { // GIVEN
	ass := assert.New(t)
	userOne := factory.OneWekanUser()
	configBoardA := factory.OneConfigBoardWithMembers(userOne)
	factory.DeactiveMembers(configBoardA, userOne)

	// set up wekanConfig for test
	wekanConfig = factory.LibwekanConfigWith(
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

	// cf la methode initReferentiel() pour mieux comprendre
	expectedDepartements := []string{"75", "62"}
	configGenerale := factory.OneConfigBoardWithMembers(user)
	factory.AddSwimlanesWithDepartments(&configGenerale, expectedDepartements...)

	// set up wekanConfig for test
	wekanConfig = factory.LibwekanConfigWith(
		[]libwekan.ConfigBoard{configGenerale},
		[]libwekan.User{user},
	)

	// WHEN
	configForUserOne := kanbanConfigForUser(user.Username)

	actual := utils.GetKeys(configForUserOne.Departements)
	// TODO lire et comprendre les articles sur la conversion de type en Golang
	// j'ai pas réussi à transformer un `[]string` en `[]CodeDepartement`
	// ni l'inverse
	expected := make([]CodeDepartement, len(expectedDepartements))
	for i, dpt := range expectedDepartements {
		expected[i] = CodeDepartement(dpt)
	}

	// THEN
	ass.ElementsMatch(expected, actual)
}

func configShouldExactlyContains(ass *assert.Assertions, config KanbanConfig, boards ...libwekan.Board) {
	boardIDsFromConfig := utils.GetKeys(config.Boards)

	expectedBoardIDs := utils.Convert(boards, func(board libwekan.Board) libwekan.BoardID {
		return board.ID
	})
	ass.ElementsMatch(boardIDsFromConfig, expectedBoardIDs)
}

func Test_parseSwimlaneTitle(t *testing.T) {
	type args struct {
		swimlane KanbanSwimlane
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test departement avec les parenthèses", args{KanbanSwimlane{Title: "93 (Seine saint-Denis)"}}, "93"},
		{"test departement sans les parenthèses", args{KanbanSwimlane{Title: "93"}}, "93"},
		{"test departement avec un espace final", args{KanbanSwimlane{Title: "93 "}}, "93"},
		{"test région avec des espaces", args{KanbanSwimlane{Title: "Île de France"}}, "Île de France"},
		{"test région - trim", args{KanbanSwimlane{Title: " Île de France "}}, "Île de France"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, parseSwimlaneTitle(tt.args.swimlane), "parseSwimlaneTitle(%v)", tt.args.swimlane)
		})
	}
}
