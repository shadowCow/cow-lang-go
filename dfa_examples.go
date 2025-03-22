package main

const (
    StateOneName = "start"
    StateTwoName = "numeric"
    StateThreeName = "wordic"
)

func createTestDfa() Dfa {

    stateOne := DfaState{
        Name: StateOneName,
        Transitions: map[rune]string{
            '0': StateTwoName,
            '1': StateTwoName,
            '2': StateTwoName,
            '3': StateTwoName,
            '4': StateTwoName,
            '5': StateTwoName,
            '6': StateTwoName,
            '7': StateTwoName,
            '8': StateTwoName,
            '9': StateTwoName,
        },
        DefaultTransition: StateThreeName,
    }
    stateTwo := DfaState{
        Name: StateTwoName,
        Transitions: map[rune]string{
            '0': StateTwoName,
            '1': StateTwoName,
            '2': StateTwoName,
            '3': StateTwoName,
            '4': StateTwoName,
            '5': StateTwoName,
            '6': StateTwoName,
            '7': StateTwoName,
            '8': StateTwoName,
            '9': StateTwoName,
        },
        DefaultTransition: StateThreeName,
    }
    stateThree := DfaState{
        Name: StateThreeName,
        Transitions: map[rune]string{},
        DefaultTransition: StateThreeName,
    }
    states := map[string]DfaState{}
    states[stateOne.Name] = stateOne;
    states[stateTwo.Name] = stateTwo;
    states[stateThree.Name] = stateThree;

    testDfa := Dfa{
		InitialState: StateOneName,
        States: states,
    }

    return testDfa;
}
